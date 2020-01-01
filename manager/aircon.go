package manager

import (
	"fmt"
	"sync"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/jmalloc/airkit/myplace"
)

const (
	constantZoneAttempts = 3
)

// AirConManager manages the state of thermostat accessories for each zone of an
// air-conditioning unit.
type AirConManager struct {
	commands chan<- []myplace.Command

	m                    sync.Mutex
	ac                   *myplace.AirCon
	thermostats          []*accessory.Thermostat
	constantZoneAttempts int
}

// NewAirConManager returns a manager for the given air-conditioning unit.
func NewAirConManager(
	commands chan<- []myplace.Command,
	ac *myplace.AirCon,
) *AirConManager {
	m := &AirConManager{
		commands: commands,
		ac:       ac,
	}

	for _, z := range ac.Zones {
		t := newThermostat(ac, z)

		t.Thermostat.TargetTemperature.OnValueRemoteUpdate(
			func(v float64) {
				m.m.Lock()
				defer m.m.Unlock()
				m.emit()
			},
		)

		t.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(
			func(int) {
				m.m.Lock()
				defer m.m.Unlock()
				m.emit()
			},
		)

		m.thermostats = append(m.thermostats, t)
	}

	m.update(ac)

	return m
}

func newThermostat(ac *myplace.AirCon, z *myplace.Zone) *accessory.Thermostat {
	n := fmt.Sprintf("%s - %s", ac.Details.Name, z.Name)
	t := accessory.NewThermostat(
		accessory.Info{
			Name:         n,
			Manufacturer: "Advantage Air & James Harris",
			Model:        "MyAir Zone",
			SerialNumber: n,
			FirmwareRevision: fmt.Sprintf(
				"%d.%d",
				ac.Details.FirmwareMajorVersion,
				ac.Details.FirmwareMinorVersion,
			),
		},
		z.CurrentTemp,
		16,
		32,
		1,
	)

	// set the current temperature range separately.
	t.Thermostat.CurrentTemperature.SetMinValue(0)
	t.Thermostat.CurrentTemperature.SetMaxValue(100)
	t.Thermostat.CurrentTemperature.SetStepValue(0.1)

	return t
}

// Accessories returns the managed accessories.
func (m *AirConManager) Accessories() []*accessory.Accessory {
	accessories := make([]*accessory.Accessory, len(m.thermostats))

	for i, t := range m.thermostats {
		accessories[i] = t.Accessory
	}

	return accessories
}

// Update updates the accessory to represent the given state.
func (m *AirConManager) Update(s *myplace.System) {
	m.m.Lock()
	defer m.m.Unlock()

	ac := s.AirConByID[m.ac.ID]
	m.update(ac)
	m.ac = ac

	m.emit()
}

// update updates the HomeKit accessories to match the air-conditioning unit.
func (m *AirConManager) update(ac *myplace.AirCon) {
	for i, z := range ac.Zones {
		t := m.thermostats[i].Thermostat

		t.CurrentTemperature.SetValue(z.CurrentTemp)
		t.TargetTemperature.SetValue(z.TargetTemp)

		if z.State == myplace.ZoneStateClosed {
			t.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		} else if ac.Details.Power == myplace.AirConPowerOff ||
			ac.Details.Mode == myplace.AirConModeVent ||
			ac.Details.Mode == myplace.AirConModeDry {
			// unsupported modes are reported as "off"
			t.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		} else if ac.Details.Mode == myplace.AirConModeCool {
			t.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateCool)
		} else if ac.Details.Mode == myplace.AirConModeHeat {
			t.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateHeat)
		}
	}
}

// update updates the air-conditioning unit match HomeKit.
func (m *AirConManager) emit() {
	var commands []myplace.Command

	defer func() {
		if len(commands) > 0 {
			m.commands <- commands
		}
	}()

	for i, z := range m.ac.Zones {
		t := m.thermostats[i].Thermostat
		target := t.TargetTemperature.GetValue()
		if z.TargetTemp != target {
			commands = append(commands, myplace.SetZoneTargetTemp(m.ac.ID, z.ID, target))
		}
	}

	power, mode := m.targetMode()

	if power != m.ac.Details.Power {
		commands = append(commands, myplace.SetAirConPower(m.ac.ID, power))
	}

	if power == myplace.AirConPowerOff {
		return
	}

	if mode != m.ac.Details.Mode {
		commands = append(commands, myplace.SetAirConMode(m.ac.ID, mode))
	}

	isCooling := mode == myplace.AirConModeCool
	open, closed := m.partitionZones(isCooling)

	modifiedNonConstantZones := false
	for _, z := range open {
		if z.State != myplace.ZoneStateOpen {
			modifiedNonConstantZones = true
			commands = append(commands, myplace.SetZoneState(m.ac.ID, z.ID, myplace.ZoneStateOpen))
		}
	}

	myzone := m.selectMyZone(isCooling, open)
	if m.ac.Details.MyZoneNumber != myzone.Number {
		commands = append(commands, myplace.SetMyZone(m.ac.ID, myzone.Number))
	}

	closedConstantZones := false
	for _, z := range closed {
		if m.ac.Zones[z.Number-1].State != myplace.ZoneStateClosed {
			if m.ac.IsConstantZone(z) {
				closedConstantZones = true

				if m.constantZoneAttempts >= constantZoneAttempts {
					continue
				}
			} else {
				modifiedNonConstantZones = true
			}

			commands = append(commands, myplace.SetZoneState(m.ac.ID, z.ID, myplace.ZoneStateClosed))
		}
	}

	if modifiedNonConstantZones {
		if m.constantZoneAttempts > constantZoneAttempts {
			fmt.Println("enabling closing of constant zones")
		}
		m.constantZoneAttempts = 0
	} else if closedConstantZones {
		m.constantZoneAttempts++
		if m.constantZoneAttempts == constantZoneAttempts+1 {
			fmt.Println("disabling closing of constant zones")
		}
	}
}

// targetMode returns the desired power and mode for the air-conditioner.
//
// It always favours cooling over heating. That is, if any zone requires
// cooling, the entire unit will be switched to cool and must reach temperature
// before the unit will be switched to heat.
func (m *AirConManager) targetMode() (myplace.AirConPower, myplace.AirConMode) {
	var needsHeating bool

	for _, t := range m.thermostats {
		cool, heat := allowedZoneModes(t.Thermostat)
		current := t.Thermostat.CurrentTemperature.GetValue()
		target := t.Thermostat.TargetTemperature.GetValue()

		if cool && current > target {
			return myplace.AirConPowerOn, myplace.AirConModeCool
		}

		if heat && current < target {
			if !needsHeating {
				needsHeating = true
			}
		}
	}

	if needsHeating {
		return myplace.AirConPowerOn, myplace.AirConModeHeat
	}

	return myplace.AirConPowerOff, m.ac.Details.Mode
}

// partioningZones returns two sets of zones, containing the zones that must be
// opened, and closed, respectively.
func (m *AirConManager) partitionZones(isCooling bool) (open, closed []*myplace.Zone) {
	for i, z := range m.ac.Zones {
		cool, heat := allowedZoneModes(m.thermostats[i].Thermostat)

		if (isCooling && cool) || (!isCooling && heat) {
			open = append(open, z)
		} else {
			closed = append(closed, z)
		}
	}

	return open, closed
}

// selectMyZone returns the best zone to use as the MyZone.
func (m *AirConManager) selectMyZone(isCooling bool, zones []*myplace.Zone) *myplace.Zone {
	var my *myplace.Zone
	var max float64

	for _, z := range zones {
		t := m.thermostats[z.Number-1].Thermostat

		current := t.CurrentTemperature.GetValue()
		target := t.TargetTemperature.GetValue()
		delta := current - target

		// if we're not cooling, favour the lowest delta (ie, current < target)
		if !isCooling {
			delta = -delta
		}

		if my == nil || delta > max {
			my = z
			max = delta
		}
	}

	return my
}

// allowedZoneModes returns booleans indicating whether a thermostat allows a
// zone to be heated and/or cooled.
func allowedZoneModes(t *service.Thermostat) (cool, heat bool) {
	switch t.TargetHeatingCoolingState.GetValue() {
	case characteristic.TargetHeatingCoolingStateCool:
		return true, false
	case characteristic.TargetHeatingCoolingStateHeat:
		return false, true
	case characteristic.TargetHeatingCoolingStateAuto:
		return true, true
	default: // characteristic.TargetHeatingCoolingStateOff:
		return false, false
	}
}
