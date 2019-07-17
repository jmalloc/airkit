package myplace

// AirConPower is an enumeration of the states of an air-conditioning unit.
type AirConPower string

const (
	// AirConPowerOn means the air-conditioning unit is turned on.
	AirConPowerOn AirConPower = "on"

	// AirConPowerOff means the air-conditioning unit is turned off.
	AirConPowerOff AirConPower = "off"
)

// AirConMode is an enumeration the operating modes of an air-conditioning unit.
type AirConMode string

const (
	// AirConModeHeat is a mode that heats the circulated air.
	AirConModeHeat AirConMode = "heat"

	// AirConModeCool is a mode that cools the circulated air.
	AirConModeCool AirConMode = "cool"

	// AirConModeVent is a mode that circulates air without heating or cooling it.
	AirConModeVent AirConMode = "vent"

	// AirConModeDry is a mode that lowers the humidity of circulated air.
	AirConModeDry AirConMode = "dry"
)

// FanSpeed is an enumeration of fan speeds of an air-conditioning unit.
type FanSpeed string

const (
	// FanSpeedLow is the slowest fan speed setting.
	FanSpeedLow FanSpeed = "low"

	// FanSpeedMedium is the medium fan speed setting.
	FanSpeedMedium FanSpeed = "medium"

	// FanSpeedHigh is the fastest fan speed setting.
	FanSpeedHigh FanSpeed = "high"

	// FanSpeedAuto is means the fan speed is automatically adjusted.
	FanSpeedAuto FanSpeed = "auto"
)

// AirCon is a ducted air-conditioning unit.
type AirCon struct {
	ID      string `json:"-"`
	Details struct {
		Name                 string      `json:"name,omitempty"`
		FanSpeed             FanSpeed    `json:"fan,omitempty"`
		Mode                 AirConMode  `json:"mode,omitempty"`
		Power                AirConPower `json:"state,omitempty"`
		FilterStatus         int         `json:"filterCleanStatus,omitempty"`
		MyZoneNumber         int         `json:"myZone,omitempty"`
		ConstantZoneNumber   int         `json:"constant1,omitempty"`
		FirmwareMajorVersion int         `json:"cbFWRevMajor,omitempty"`
		FirmwareMinorVersion int         `json:"cbFWRevMinor,omitempty"`
	} `json:"info,omitempty"`
	ZoneByID map[string]*Zone `json:"zones,omitempty"`
	Zones    []*Zone          `json:"-"`
}

func (ac *AirCon) populate(id string) {
	ac.ID = id
	ac.Zones = make([]*Zone, len(ac.ZoneByID))

	for id, z := range ac.ZoneByID {
		z.populate(id)
		ac.Zones[z.Number-1] = z
	}
}

// SetAirConPower returns a command that turns and air-conditioning unit on or
// off.
func SetAirConPower(id string, v AirConPower) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[id]

		if !ok {
			ac = &AirCon{}
			req[id] = ac
		}

		ac.Details.Power = v
	}
}

// SetAirConMode returns a command that sets the mode of an air-conditioning unit.
func SetAirConMode(id string, v AirConMode) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[id]

		if !ok {
			ac = &AirCon{}
			req[id] = ac
		}

		ac.Details.Mode = v
	}
}

// SetFanSpeed returns a command that sets the fan mode of an air-conditioning unit.
func SetFanSpeed(id string, v FanSpeed) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[id]

		if !ok {
			ac = &AirCon{}
			req[id] = ac
		}

		ac.Details.FanSpeed = v
	}
}
