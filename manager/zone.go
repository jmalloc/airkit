package manager

import (
	"fmt"
	"log"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/jmalloc/airkit/myplace"
)

// ZoneManager manages the state of a zone.
type ZoneManager struct {
	commands   chan<- myplace.Command
	ac         *myplace.AirCon
	z          *myplace.Zone
	accessory  *accessory.Accessory
	thermostat *service.Thermostat
}

// NewZoneManager returns a manager for the given zone.
func NewZoneManager(
	commands chan<- myplace.Command,
	ac *myplace.AirCon,
	z *myplace.Zone,
) *ZoneManager {
	n := fmt.Sprintf("%s - %s", ac.Details.Name, z.Name)

	m := &ZoneManager{
		commands: commands,
		ac:       ac,
		z:        z,
		accessory: accessory.New(
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
			accessory.TypeThermostat,
		),
		thermostat: service.NewThermostat(),
	}

	m.accessory.AddService(m.thermostat.Service)

	m.thermostat.CurrentTemperature.SetEventsEnabled(true)
	m.thermostat.CurrentTemperature.SetMinValue(0)
	m.thermostat.CurrentTemperature.SetMaxValue(100)
	m.thermostat.CurrentTemperature.SetStepValue(0.5)

	m.thermostat.CurrentHeatingCoolingState.SetEventsEnabled(true)

	m.thermostat.TargetTemperature.SetEventsEnabled(true)
	m.thermostat.TargetTemperature.SetMinValue(16)
	m.thermostat.TargetTemperature.SetMaxValue(32)
	m.thermostat.TargetTemperature.SetStepValue(1)
	m.thermostat.TargetTemperature.OnValueRemoteUpdate(m.setTargetTemp)

	m.thermostat.TargetHeatingCoolingState.SetEventsEnabled(true)
	m.thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(m.setTargetState)

	m.update(ac, z)

	return m
}

// Accessories returns the managed accessories.
func (m *ZoneManager) Accessories() []*accessory.Accessory {
	return []*accessory.Accessory{
		m.accessory,
	}
}

// Update updates the accessory to represent the given state.
func (m *ZoneManager) Update(s *myplace.System) {
	ac := s.AirConByID[m.ac.ID]
	z := ac.ZoneByID[m.z.ID]
	m.update(ac, z)
}

func (m *ZoneManager) update(ac *myplace.AirCon, z *myplace.Zone) {
	m.thermostat.CurrentTemperature.SetValue(z.CurrentTemp)
	m.thermostat.TargetTemperature.SetValue(z.TargetTemp)

	if z.State == myplace.ZoneStateClosed {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateOff)
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		return
	}

	if ac.Details.MyZoneNumber != z.Number {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateAuto)
	} else if ac.Details.Mode == myplace.AirConModeCool {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateCool)
	} else if ac.Details.Mode == myplace.AirConModeHeat {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateHeat)
	} else {
		// unsupported modes are treated equivalent to "off"
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateOff)
	}

	if ac.Details.Power == myplace.AirConPowerOff ||
		ac.Details.Mode == myplace.AirConModeVent ||
		ac.Details.Mode == myplace.AirConModeDry {
		// unsupported modes are treated equivalent to "off"
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
	} else if ac.Details.Mode == myplace.AirConModeCool {
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateCool)
	} else if ac.Details.Mode == myplace.AirConModeHeat {
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateHeat)
	}

	m.ac = ac
	m.z = z
}

func (m *ZoneManager) setTargetTemp(v float64) {
	log.Printf("%s/%s target temp = %0.1f%%", m.ac.ID, m.z.ID, v)

	m.commands <- myplace.SetZoneTargetTemp(m.ac.ID, m.z.ID, v)
}

func (m *ZoneManager) setTargetState(v int) {
	log.Printf("%s/%s target state = %d", m.ac.ID, m.z.ID, v)

	switch v {
	case characteristic.TargetHeatingCoolingStateOff:
		if m.ac.Details.MyZoneNumber == m.z.Number {
			m.commands <- myplace.SetMyZone(m.ac.ID, m.ac.Details.ConstantZoneNumber)
		}
		m.commands <- myplace.SetZoneState(m.ac.ID, m.z.ID, myplace.ZoneStateClosed)

	case characteristic.TargetHeatingCoolingStateCool:
		m.commands <- myplace.SetMyZone(m.ac.ID, m.z.Number)
		m.commands <- myplace.SetAirConMode(m.ac.ID, myplace.AirConModeCool)
		m.commands <- myplace.SetZoneState(m.ac.ID, m.z.ID, myplace.ZoneStateOpen)

	case characteristic.TargetHeatingCoolingStateHeat:
		m.commands <- myplace.SetMyZone(m.ac.ID, m.z.Number)
		m.commands <- myplace.SetAirConMode(m.ac.ID, myplace.AirConModeHeat)
		m.commands <- myplace.SetZoneState(m.ac.ID, m.z.ID, myplace.ZoneStateOpen)

	case characteristic.TargetHeatingCoolingStateAuto:
		m.commands <- myplace.SetZoneState(m.ac.ID, m.z.ID, myplace.ZoneStateOpen)
	}
}
