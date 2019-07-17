package manager

import (
	"fmt"

	"github.com/brutella/hc/characteristic"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	"github.com/jmalloc/airkit/myplace"
)

// AirConManager manages the stateo of an air-conditioning unit accessory.
type AirConManager struct {
	commands     chan<- myplace.Command
	id           string
	acAccessory  *accessory.Accessory
	fanAccessory *accessory.Accessory
	power        *service.Switch
	fan          *service.FanV2
	speed        *characteristic.RotationSpeed
	prevSpeed    myplace.FanSpeed
}

// NewAirConManager returns a manager for the given air-conditioning unit.
func NewAirConManager(
	commands chan<- myplace.Command,
	ac *myplace.AirCon,
) *AirConManager {
	m := &AirConManager{
		commands: commands,
		id:       ac.ID,
		acAccessory: accessory.New(
			accessory.Info{
				Name:         ac.Details.Name,
				Manufacturer: "Advantage Air & James Harris",
				Model:        "MyAir Air Conditioner",
				SerialNumber: ac.ID,
				FirmwareRevision: fmt.Sprintf(
					"%d.%d",
					ac.Details.FirmwareMajorVersion,
					ac.Details.FirmwareMinorVersion,
				),
			},
			accessory.TypeAirConditioner,
		),
		fanAccessory: accessory.New(
			accessory.Info{
				Name:         ac.Details.Name + " Fan Speed Override",
				Manufacturer: "Advantage Air & James Harris",
				Model:        "MyAir Air Conditioner Fan Speed Override",
				SerialNumber: ac.ID,
				FirmwareRevision: fmt.Sprintf(
					"%d.%d",
					ac.Details.FirmwareMajorVersion,
					ac.Details.FirmwareMinorVersion,
				),
			},
			accessory.TypeAirConditioner,
		),
		power:     service.NewSwitch(),
		fan:       service.NewFanV2(),
		speed:     characteristic.NewRotationSpeed(),
		prevSpeed: myplace.FanSpeedMedium,
	}

	m.acAccessory.AddService(m.power.Service)
	m.power.On.SetEventsEnabled(true)
	m.power.On.OnValueRemoteUpdate(m.setPower)

	m.fanAccessory.AddService(m.fan.Service)
	m.fan.Active.SetEventsEnabled(true)
	m.fan.Active.OnValueRemoteUpdate(m.setFanActive)

	m.fan.AddCharacteristic(m.speed.Characteristic)
	m.speed.SetEventsEnabled(true)
	m.speed.OnValueRemoteUpdate(m.setFanSpeed)

	m.update(ac)

	return m
}

// Accessories returns the managed accessories.
func (m *AirConManager) Accessories() []*accessory.Accessory {
	return []*accessory.Accessory{
		m.acAccessory,
		m.fanAccessory,
	}
}

// Update updates the accessory to represent the given state.
func (m *AirConManager) Update(s *myplace.System) {
	ac := s.AirConByID[m.id]
	m.update(ac)
}

func (m *AirConManager) update(ac *myplace.AirCon) {
	m.power.On.SetValue(ac.Details.Power == myplace.AirConPowerOn)

	if ac.Details.FanSpeed == myplace.FanSpeedAuto {
		m.fan.Active.SetValue(characteristic.ActiveInactive)
	} else {
		m.prevSpeed = ac.Details.FanSpeed
		m.fan.Active.SetValue(characteristic.ActiveActive)
		m.speed.SetValue(marshalFanSpeed(ac.Details.FanSpeed))
	}
}

func (m *AirConManager) setPower(v bool) {
	if v {
		m.commands <- myplace.SetAirConPower(m.id, myplace.AirConPowerOn)
	} else {
		m.commands <- myplace.SetAirConPower(m.id, myplace.AirConPowerOff)
	}
}

func (m *AirConManager) setFanActive(v int) {
	switch v {
	case characteristic.ActiveActive:
		m.commands <- myplace.SetFanSpeed(m.id, m.prevSpeed)
	case characteristic.ActiveInactive:
		m.commands <- myplace.SetFanSpeed(m.id, myplace.FanSpeedAuto)
	}
}

func (m *AirConManager) setFanSpeed(v float64) {
	m.commands <- myplace.SetFanSpeed(m.id, unmarshalFanSpeed(v))
}

func marshalFanSpeed(v myplace.FanSpeed) float64 {
	switch v {
	case myplace.FanSpeedHigh:
		return 100
	case myplace.FanSpeedMedium:
		return 50
	case myplace.FanSpeedLow:
		return 25
	default:
		return 0
	}
}

func unmarshalFanSpeed(v float64) myplace.FanSpeed {
	if v == 0 {
		return myplace.FanSpeedAuto
	} else if v <= 33.333 {
		return myplace.FanSpeedLow
	} else if v <= 66.666 {
		return myplace.FanSpeedMedium
	} else {
		return myplace.FanSpeedHigh
	}
}
