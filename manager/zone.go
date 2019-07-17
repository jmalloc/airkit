package manager

import (
	"fmt"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/jmalloc/airkit/myplace"
)

// ZoneManager manages the state of a zone.
type ZoneManager struct {
	commands   chan<- myplace.Command
	acID       string
	zoneID     string
	isMyZone   bool
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
		acID:     ac.ID,
		zoneID:   z.ID,
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
	ac := s.AirConByID[m.acID]
	z := ac.ZoneByID[m.zoneID]
	m.update(ac, z)
}

func (m *ZoneManager) update(ac *myplace.AirCon, z *myplace.Zone) {
	m.isMyZone = ac.Details.MyZoneNumber == z.Number

	m.thermostat.CurrentTemperature.SetValue(z.CurrentTemp)
	m.thermostat.TargetTemperature.SetValue(z.TargetTemp)

	if z.State == myplace.ZoneStateClosed {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateOff)
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		return
	}

	// unsupported modes are treated equivalent to "off"
	if ac.Details.Power == myplace.AirConPowerOff ||
		ac.Details.Mode == myplace.AirConModeVent ||
		ac.Details.Mode == myplace.AirConModeDry {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateAuto)
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateOff)
		return
	}

	if ac.Details.Mode == myplace.AirConModeCool {
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateCool)
	} else if ac.Details.Mode == myplace.AirConModeHeat {
		m.thermostat.CurrentHeatingCoolingState.SetValue(characteristic.CurrentHeatingCoolingStateHeat)
	}

	if !m.isMyZone {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateAuto)
	} else if ac.Details.Mode == myplace.AirConModeCool {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateCool)
	} else if ac.Details.Mode == myplace.AirConModeHeat {
		m.thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateHeat)
	}
}

func (m *ZoneManager) setTargetTemp(v float64) {
	// 	log.Printf("--------- %s target temperature set to %f degrees",
	// 		m.accessory.Info.Name.GetValue(),
	// 		v,
	// 	)
	// 	// m.poller.Send(
	// 	// 	api.UpdateTargetTemperature(
	// 	// 		m.acKey,
	// 	// 		m.zoneKey,
	// 	// 		v,
	// 	// 	),
	// 	// )
}

func (m *ZoneManager) setTargetState(v int) {
	// 	switch v {
	// 	case characteristic.TargetHeatingCoolingStateOff:
	// 		log.Printf("--------- %s turned off",
	// 			m.accessory.Info.Name.GetValue(),
	// 		)

	// 		// if m.isMyZone {
	// 		// 	m.poller.Send(
	// 		// 		api.UpdateAirConState(
	// 		// 			m.acKey,
	// 		// 			api.Off,
	// 		// 		),
	// 		// 	)
	// 		// } else {
	// 		// 	m.poller.Send(
	// 		// 		api.UpdateZoneState(
	// 		// 			m.acKey,
	// 		// 			m.zoneKey,
	// 		// 			api.Closed,
	// 		// 		),
	// 		// 	)
	// }

	// 	case characteristic.TargetHeatingCoolingStateCool:
	// 		log.Printf("--------- %s set to cool",
	// 			m.accessory.Info.Name.GetValue(),
	// 		)
	// 		// m.poller.Send(
	// 		// 	api.UpdateModeAndMyZone(m.acKey, api.Cool, m.number),
	// 		// )

	// 	case characteristic.TargetHeatingCoolingStateHeat:
	// 		log.Printf("--------- %s set to heat",
	// 			m.accessory.Info.Name.GetValue(),
	// 		)
	// 		// m.poller.Send(
	// 		// 	api.UpdateModeAndMyZone(m.acKey, api.Heat, m.number),
	// 		// )

	// 	case characteristic.TargetHeatingCoolingStateAuto:
	// 		log.Printf("--------- %s set to auto",
	// 			m.accessory.Info.Name.GetValue(),
	// 		)
	// 		// m.poller.Send(
	// 		// 	api.UpdateZoneState(
	// 		// 		m.acKey,
	// 		// 		m.zoneKey,
	// 		// 		api.Open,
	// 		// 	),
	// 		// )
	// 	}
}
