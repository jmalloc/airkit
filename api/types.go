package api

// System describes a system consisting of multiple air-conditioning units.
type System struct {
	AirCons map[string]AirCon `json:"aircons"`

	Details struct {
		AppVersion       string `json:"myAppRev"`
		ServiceVersion   string `json:"aaServiceRev"`
		NeedsUpdate      bool   `json:"needsUpdate"`
		TouchScreenID    string `json:"rid"`
		TouchScreenModel string `json:"tspModel"`
	} `json:"system"`
}

// AirConState is an enumeration of the states of an air-conditioning unit.
type AirConState string

const (
	// On means the air-conditioning unit is turned on.
	On AirConState = "on"

	// Off means the air-conditioning unit is turned off.
	Off AirConState = "off"
)

// FanSpeed is an enumeration of fan speeds of an air-conditioning unit.
type FanSpeed string

const (
	// Low is the slowest fan speed setting.
	Low FanSpeed = "low"

	// Medium is the medium fan speed setting.
	Medium FanSpeed = "medium"

	// High is the fastest fan speed setting.
	High FanSpeed = "high"

	// Auto is means the fan speed is automatically adjusted.
	Auto FanSpeed = "auto"
)

// Mode is an enumeration the operating modes of an air-conditioning unit.
type Mode string

const (
	// Heat is a mode that heats the circulated air.
	Heat Mode = "heat"

	// Cool is a mode that cools the circulated air.
	Cool Mode = "cool"

	// Vent is a mode that circulates air without heating or cooling it.
	Vent Mode = "vent"

	// Dry is a mode that lowers the humidity of circulated air.
	Dry Mode = "dry"
)

// AirCon contains information about an air-conditioning unit.
type AirCon struct {
	Details struct {
		Name         string      `json:"name"`
		FanSpeed     FanSpeed    `json:"fan"`
		Mode         Mode        `json:"mode"`
		State        AirConState `json:"state"`
		MyZoneNumber int         `json:"myZone"`
	} `json:"info"`

	Zones map[string]Zone `json:"zones"`
}

// MyZone returns the currently select "MyZone"
func (ac AirCon) MyZone() Zone {
	for _, z := range ac.Zones {
		if z.Number == ac.Details.MyZoneNumber {
			return z
		}
	}

	panic("no zone selected")
}

// IsOn returns true if this unit is turned on.
func (ac AirCon) IsOn() bool {
	return ac.Details.State == On
}

// ZoneState is an enumeration of the states of a zone.
type ZoneState string

const (
	// Open means the damper for this zone is open.
	Open ZoneState = "open"

	// Closed means the damper for this zone is closed.
	Closed ZoneState = "close" // note: value is 'close', without the trailing 'd'.
)

// Zone is a vent or collection of vents connected to a ducted air-conditioning
// unit.
type Zone struct {
	Number           int       `json:"number"`
	Name             string    `json:"name"`
	State            ZoneState `json:"state"`
	HasTempControl   int       `json:"type"`
	TargetTemp       float64   `json:"setTemp"`
	DamperPercentage int       `json:"value"` // 5 - 1000
	CurrentTemp      float64   `json:"measuredTemp"`
}

// IsOn returns true if this zone is turned on.
func (z Zone) IsOn() bool {
	return z.State == Open
}
