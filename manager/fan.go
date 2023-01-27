package manager

import (
	"fmt"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
	"github.com/brutella/hap/service"
	"github.com/jmalloc/airkit/myplace"
)

// FanManager manages the state of a "fan speed" accessory.
//
// This allows the user to set an air-conditioning unit's fan speed via HomeKit.
//
// Note that because HomeKit does not have a fan accessory with an "auto"
// setting, the accessory is presented as an "override" to the auto setting.  If
// the AC unit is "on", and the fan accessory is "off", the fan speed is set to
// auto.
//
// I've found this provides the best experience when using Siri to control the
// fan speed, allowing phrases like "Turn off the fan speed override".  I
// couldn't work out any combination of characteristics that would allow phrases
// like "Set the fan speed to auto", which would be ideal.
type FanManager struct {
	commands  chan<- []myplace.Command
	acID      string
	autoSpeed myplace.FanSpeed
	prevSpeed myplace.FanSpeed
	accessory *accessory.A
	fan       *service.FanV2
	speed     *characteristic.RotationSpeed
}

// NewFanManager returns a manager for the given air-conditioning unit's fan.
func NewFanManager(
	commands chan<- []myplace.Command,
	ac *myplace.AirCon,
) *FanManager {
	m := &FanManager{
		commands: commands,
		acID:     ac.ID,
		accessory: accessory.New(
			accessory.Info{
				Name:         ac.Details.Name + " Fan Speed Override",
				Manufacturer: "Advantage Air & James Harris",
				Model:        "MyAir Air Conditioner Fan Speed Override",
				SerialNumber: ac.ID,
				Firmware: fmt.Sprintf(
					"%d.%d",
					ac.Details.FirmwareMajorVersion,
					ac.Details.FirmwareMinorVersion,
				),
			},
			accessory.TypeAirConditioner,
		),
		fan:       service.NewFanV2(),
		speed:     characteristic.NewRotationSpeed(),
		prevSpeed: myplace.FanSpeedMedium,
	}

	m.accessory.AddS(m.fan.S)
	m.fan.Active.OnValueRemoteUpdate(m.setFanActive)

	m.fan.AddC(m.speed.C)
	m.speed.OnValueRemoteUpdate(m.setFanSpeed)

	m.update(ac)

	return m
}

// Accessories returns the managed accessories.
func (m *FanManager) Accessories() []*accessory.A {
	return []*accessory.A{
		m.accessory,
	}
}

// Update updates the accessory to represent the given state.
func (m *FanManager) Update(s *myplace.System) {
	ac := s.AirConByID[m.acID]
	m.update(ac)
}

func (m *FanManager) update(ac *myplace.AirCon) {
	switch ac.Details.FanSpeed {
	case myplace.FanSpeedAutoHardware, myplace.FanSpeedAutoSoftware:
		m.fan.Active.SetValue(characteristic.ActiveInactive)
	default:
		m.prevSpeed = ac.Details.FanSpeed
		m.fan.Active.SetValue(characteristic.ActiveActive)
		m.speed.SetValue(m.marshalFanSpeed(ac.Details.FanSpeed))
	}

	if ac.Details.MyFanEnabled {
		m.autoSpeed = myplace.FanSpeedAutoSoftware
	} else {
		m.autoSpeed = myplace.FanSpeedAutoHardware
	}
}

func (m *FanManager) setFanActive(v int) {
	switch v {
	case characteristic.ActiveActive:
		m.commands <- []myplace.Command{myplace.SetFanSpeed(m.acID, m.prevSpeed)}
	case characteristic.ActiveInactive:
		m.commands <- []myplace.Command{myplace.SetFanSpeed(m.acID, m.autoSpeed)}
	}
}

func (m *FanManager) setFanSpeed(v float64) {
	m.commands <- []myplace.Command{myplace.SetFanSpeed(m.acID, m.unmarshalFanSpeed(v))}
}

func (m *FanManager) marshalFanSpeed(v myplace.FanSpeed) float64 {
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

func (m *FanManager) unmarshalFanSpeed(v float64) myplace.FanSpeed {
	if v == 0 {
		return m.autoSpeed
	} else if v <= 33.333 {
		return myplace.FanSpeedLow
	} else if v <= 66.666 {
		return myplace.FanSpeedMedium
	} else {
		return myplace.FanSpeedHigh
	}
}
