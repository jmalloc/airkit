package myplace

// ZoneState is an enumeration of the states of a zone.
type ZoneState string

const (
	// ZoneStateOpen means the damper for this zone is open.
	ZoneStateOpen ZoneState = "open"

	// ZoneStateClosed means the damper for this zone is closed.
	ZoneStateClosed ZoneState = "close" // note: value is 'close', without the trailing 'd'.
)

// Zone is a vent or collection of vents connected to a ducted air-conditioning
// unit.
type Zone struct {
	ID               string    `json:"-"`
	Number           int       `json:"number,omitempty"`
	Name             string    `json:"name,omitempty"`
	State            ZoneState `json:"state,omitempty"`
	DamperPercentage int       `json:"value,omitempty"` // 5 - 1000
	HasTempControl   int       `json:"type,omitempty"`
	CurrentTemp      float64   `json:"measuredTemp,omitempty"`
	TargetTemp       float64   `json:"setTemp,omitempty"`
}

func (z *Zone) populate(id string) {
	z.ID = id
}
