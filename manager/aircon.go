package manager

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
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
	zoneAccessories      []*zoneAccessories
	constantZoneAttempts int
}

type zoneAccessories struct {
	Accessories     []*accessory.A
	Thermostat      *service.Thermostat
	Battery         *characteristic.StatusLowBattery
	MyZoneIndicator *service.ContactSensor
}

// NewAirConManager returns a manager for the given air-conditioning unit.
func NewAirConManager(
	store hap.Store,
	commands chan<- []myplace.Command,
	ac *myplace.AirCon,
) *AirConManager {
	m := &AirConManager{
		commands: commands,
		ac:       ac,
	}

	for _, z := range ac.Zones {
		a := newZoneAccessories(ac, z)

		a.Thermostat.TargetTemperature.OnValueRemoteUpdate(
			func(v float64) {
				m.m.Lock()
				defer m.m.Unlock()
				m.apply()
			},
		)

		key := fmt.Sprintf("myplace-%s-%s-target-state", ac.ID, z.ID)
		if v, err := store.Get(key); err == nil {
			if i, err := strconv.Atoi(string(v)); err == nil {
				a.Thermostat.TargetHeatingCoolingState.SetValue(i)
			}
		}

		a.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(
			func(v int) {
				store.Set(key, []byte(strconv.Itoa(v)))

				m.m.Lock()
				defer m.m.Unlock()
				m.apply()
			},
		)

		m.zoneAccessories = append(m.zoneAccessories, a)
	}

	m.update(ac)

	return m
}

func newZoneAccessories(ac *myplace.AirCon, z *myplace.Zone) *zoneAccessories {
	n := fmt.Sprintf("%s %s", z.Name, ac.Details.Name)
	t := accessory.NewThermostat(
		accessory.Info{
			Name:         n,
			Manufacturer: "Advantage Air & James Harris",
			Model:        "MyAir Zone",
			SerialNumber: fmt.Sprintf("%s.%s", ac.ID, z.ID),
			Firmware: fmt.Sprintf(
				"%d.%d",
				ac.Details.FirmwareMajorVersion,
				ac.Details.FirmwareMinorVersion,
			),
		},
	)
	t.Id = makeZoneAccessoryID(ac, z, zoneThermostatID)

	t.Thermostat.TargetTemperature.SetMinValue(16)
	t.Thermostat.TargetTemperature.SetMaxValue(32)
	t.Thermostat.TargetTemperature.SetStepValue(1)

	t.Thermostat.CurrentTemperature.SetMinValue(0)
	t.Thermostat.CurrentTemperature.SetMaxValue(100)
	t.Thermostat.CurrentTemperature.SetStepValue(0.1)

	b := characteristic.NewStatusLowBattery()
	t.Thermostat.AddC(b.C)

	m := accessory.New(
		accessory.Info{
			Name:         fmt.Sprintf("%s MyZone", z.Name),
			Manufacturer: "Advantage Air & James Harris",
			Model:        "MyAir Zone",
			SerialNumber: fmt.Sprintf("%s.%s", ac.ID, z.ID),
			Firmware: fmt.Sprintf(
				"%d.%d",
				ac.Details.FirmwareMajorVersion,
				ac.Details.FirmwareMinorVersion,
			),
		},
		accessory.TypeSensor,
	)
	m.Id = makeZoneAccessoryID(ac, z, zoneMyZoneIndicatorID)

	cs := service.NewContactSensor()
	m.AddS(cs.S)

	return &zoneAccessories{
		Accessories:     []*accessory.A{t.A, m},
		Thermostat:      t.Thermostat,
		Battery:         b,
		MyZoneIndicator: cs,
	}
}

// Accessories returns the managed accessories.
func (m *AirConManager) Accessories() []*accessory.A {
	var accessories []*accessory.A

	for _, a := range m.zoneAccessories {
		accessories = append(accessories, a.Accessories...)
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

	m.apply()
}

// update updates the HomeKit accessories to match the air-conditioning unit.
func (m *AirConManager) update(ac *myplace.AirCon) {
	for i, z := range ac.Zones {
		a := m.zoneAccessories[i]

		a.Thermostat.CurrentTemperature.SetValue(z.CurrentTemp)
		a.Thermostat.TargetTemperature.SetValue(z.TargetTemp)

		if z.State == myplace.ZoneStateClosed {
			a.Thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		} else if ac.Details.Power == myplace.AirConPowerOff ||
			ac.Details.Mode == myplace.AirConModeVent ||
			ac.Details.Mode == myplace.AirConModeDry {
			// unsupported modes are reported as "off"
			a.Thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		} else if ac.Details.Mode == myplace.AirConModeCool {
			a.Thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateCool)
		} else if ac.Details.Mode == myplace.AirConModeHeat {
			a.Thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateHeat)
		}

		if z.Error == myplace.ZoneErrorNone {
			a.Battery.SetValue(characteristic.StatusLowBatteryBatteryLevelNormal)
		} else {
			a.Battery.SetValue(characteristic.StatusLowBatteryBatteryLevelLow)
		}

		if z.Number == ac.Details.MyZoneNumber {
			a.MyZoneIndicator.ContactSensorState.SetValue(characteristic.ContactSensorStateContactDetected)
		} else {
			a.MyZoneIndicator.ContactSensorState.SetValue(characteristic.ContactSensorStateContactNotDetected)
		}
	}
}

// apply updates the air-conditioning unit to match HomeKit.
func (m *AirConManager) apply() {
	var commands []myplace.Command

	defer func() {
		if len(commands) > 0 {
			m.commands <- commands
		}
	}()

	for i, z := range m.ac.Zones {
		t := m.zoneAccessories[i].Thermostat
		target := t.TargetTemperature.Value()
		if z.TargetTemp != target {
			commands = append(commands, myplace.SetZoneTargetTemp(m.ac.ID, z, target))
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
			commands = append(commands, myplace.SetZoneState(m.ac.ID, z, myplace.ZoneStateOpen))
		}
	}

	if z, ok := m.selectMyZone(isCooling, open); ok {
		if m.ac.Details.MyZoneNumber != z.Number {
			commands = append(commands, myplace.SetMyZone(m.ac.ID, z))
		}
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

			commands = append(commands, myplace.SetZoneState(m.ac.ID, z, myplace.ZoneStateClosed))
		}
	}

	if modifiedNonConstantZones {
		if m.constantZoneAttempts > constantZoneAttempts {
			log.Print("enabling closing of constant zones")
		}
		m.constantZoneAttempts = 0
	} else if closedConstantZones {
		m.constantZoneAttempts++
		if m.constantZoneAttempts == constantZoneAttempts+1 {
			log.Print("disabling closing of constant zones")
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

	// Cool until we're a little below the target temperature. This is an
	// attempt to let the AC regulate the temperature. That is, we only switch
	// the unit off if the AC is over-cooling.
	const coolThreshold = -0.1

	// Don't start heating until we're below the target temperature. This is an
	// attempt avoid continually switching between heating and cooling when zones
	// are set to AUTO. In general, we favour cooling over heating.
	const heatThreshold = -0.5

	for _, a := range m.zoneAccessories {
		cool, heat := allowedZoneModes(a.Thermostat)
		current := a.Thermostat.CurrentTemperature.Value()
		target := a.Thermostat.TargetTemperature.Value()
		delta := current - target

		if cool && delta > coolThreshold {
			return myplace.AirConPowerOn, myplace.AirConModeCool
		}

		if heat && delta < heatThreshold {
			needsHeating = true
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
		cool, heat := allowedZoneModes(m.zoneAccessories[i].Thermostat)

		if (isCooling && cool) || (!isCooling && heat) {
			open = append(open, z)
		} else {
			closed = append(closed, z)
		}
	}

	return open, closed
}

// selectMyZone returns the best zone to use as the MyZone.
func (m *AirConManager) selectMyZone(isCooling bool, zones []*myplace.Zone) (*myplace.Zone, bool) {
	var my *myplace.Zone
	var max float64

	for _, z := range zones {
		// Don't consider the zone a candidate for MyZone if we don't even want
		// it on.
		if z.State != myplace.ZoneStateOpen {
			continue
		}

		// Don't consider the zone a candidate for MyZone if we can't measure
		// the temperature.
		if z.Error == myplace.ZoneErrorNoSignal {
			continue
		}

		t := m.zoneAccessories[z.Number-1].Thermostat

		current := t.CurrentTemperature.Value()
		target := t.TargetTemperature.Value()
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

	return my, my != nil
}

// allowedZoneModes returns booleans indicating whether a thermostat allows a
// zone to be heated and/or cooled.
func allowedZoneModes(t *service.Thermostat) (cool, heat bool) {
	switch t.TargetHeatingCoolingState.Value() {
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
