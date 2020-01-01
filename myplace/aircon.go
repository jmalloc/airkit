package myplace

// AirConPower is an enumeration of the states of an air-conditioning unit.
type AirConPower string

const (
	// AirConPowerOn means the air-conditioning unit is turned on.
	AirConPowerOn AirConPower = "on"

	// AirConPowerOff means the air-conditioning unit is turned off.
	AirConPowerOff AirConPower = "off"
)

func (p AirConPower) String() string {
	switch p {
	case AirConPowerOn:
		return "on"
	case AirConPowerOff:
		return "off"
	default:
		return "unknown"
	}
}

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

	// AirConModeAuto is a mode that sets the actual mode automatically.
	AirConModeAuto AirConMode = "myauto"
)

func (m AirConMode) String() string {
	switch m {
	case AirConModeHeat:
		return "heat"
	case AirConModeCool:
		return "cool"
	case AirConModeVent:
		return "fan" // fan is what's shown in the MyPlace app.
	case AirConModeDry:
		return "dry"
	case AirConModeAuto:
		return "auto"
	default:
		return "unknown"
	}
}

// FanSpeed is an enumeration of fan speeds of an air-conditioning unit.
type FanSpeed string

const (
	// FanSpeedLow is the slowest fan speed setting.
	FanSpeedLow FanSpeed = "low"

	// FanSpeedMedium is the medium fan speed setting.
	FanSpeedMedium FanSpeed = "medium"

	// FanSpeedHigh is the fastest fan speed setting.
	FanSpeedHigh FanSpeed = "high"

	// FanSpeedAutoHardware is means the fan speed is automatically adjusted by the
	// air-conditioning unit. This option is only usable if the "MyFan" feature
	// is DISABLED.
	FanSpeedAutoHardware FanSpeed = "auto"

	// FanSpeedAutoSoftware is means the fan speed is automatically adjusted by
	// the MyPlace software. This option is only usable if the "MyFan" feature
	// is ENABLED.
	FanSpeedAutoSoftware FanSpeed = "autoAA"
)

func (s FanSpeed) String() string {
	switch s {
	case FanSpeedLow:
		return "low"
	case FanSpeedMedium:
		return "medium"
	case FanSpeedHigh:
		return "high"
	case FanSpeedAutoHardware, FanSpeedAutoSoftware:
		return "auto"
	default:
		return "unknown"
	}
}

// AirCon is a ducted air-conditioning unit.
type AirCon struct {
	ID      string `json:"-"`
	Details struct {
		Name                 string      `json:"name,omitempty"`
		FanSpeed             FanSpeed    `json:"fan,omitempty"`
		Mode                 AirConMode  `json:"mode,omitempty"`
		Power                AirConPower `json:"state,omitempty"`
		FilterStatus         int         `json:"filterCleanStatus,omitempty"`
		MyFanEnabled         bool        `json:"aaAutoFanModeEnabled,omitempty"`
		MyTempEnabled        bool        `json:"climateControlModeEnabled,omitempty"`
		MyTempRunning        bool        `json:"climateControlModeIsRunning,omitempty"`
		MyAutoEnabled        bool        `json:"myAutoModeEnabled,omitempty"`
		MyAutoRunning        bool        `json:"myAutoModeIsRunning,omitempty"`
		MyAutoMode           AirConMode  `json:"myAutoModeCurrentSetMode,omitempty"`
		MySleepSaverEnabled  bool        `json:"quietNightModeEnabled,omitempty"`
		MySleepSaverRunning  bool        `json:"quietNightModeIsRunning,omitempty"`
		MyZoneNumber         int         `json:"myZone,omitempty"`
		ConstantZone1Number  int         `json:"constant1,omitempty"`
		ConstantZone2Number  int         `json:"constant2,omitempty"`
		ConstantZone3Number  int         `json:"constant3,omitempty"`
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

// MyZone returns the currently selected "MyZone".
func (ac *AirCon) MyZone() *Zone {
	return ac.Zones[ac.Details.MyZoneNumber-1]
}

// IsMyZone returns true if z is the current MyZone.
func (ac *AirCon) IsMyZone(z *Zone) bool {
	return z.Number == ac.Details.MyZoneNumber
}

// ConstantZones returns the zones that are configured as "constant zones".
func (ac *AirCon) ConstantZones() []*Zone {
	var zones []*Zone

	if ac.Details.ConstantZone1Number != 0 {
		zones = append(zones, ac.Zones[ac.Details.ConstantZone1Number-1])
	}

	if ac.Details.ConstantZone2Number != 0 {
		zones = append(zones, ac.Zones[ac.Details.ConstantZone2Number-1])
	}

	if ac.Details.ConstantZone3Number != 0 {
		zones = append(zones, ac.Zones[ac.Details.ConstantZone3Number-1])
	}

	return zones
}

// IsConstantZone returns true if z is configured as a constant zone.
func (ac *AirCon) IsConstantZone(z *Zone) bool {
	switch z.Number {
	case ac.Details.ConstantZone1Number,
		ac.Details.ConstantZone2Number,
		ac.Details.ConstantZone3Number:
		return true
	default:
		return false
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
