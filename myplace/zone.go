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

// SetMyZone returns a command that sets "MyZone" of an air-conditioning unit.
func SetMyZone(id string, n int) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[id]

		if !ok {
			ac = &AirCon{}
			req[id] = ac
		}

		ac.Details.MyZoneNumber = n
	}
}

// SetZoneState returns a command that sets state of a zone.
func SetZoneState(acID, zID string, v ZoneState) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[acID]

		if !ok {
			ac = &AirCon{}
			req[acID] = ac
		}

		z, ok := ac.ZoneByID[zID]

		if !ok {
			z = &Zone{}

			if ac.ZoneByID == nil {
				ac.ZoneByID = map[string]*Zone{}
			}

			ac.ZoneByID[zID] = z
		}

		z.State = v
	}
}

// SetZoneTargetTemp returns a command that sets the fan mode of an air-conditioning unit.
func SetZoneTargetTemp(acID, zID string, v float64) Command {
	return func(req map[string]*AirCon) {
		ac, ok := req[acID]

		if !ok {
			ac = &AirCon{}
			req[acID] = ac
		}

		z, ok := ac.ZoneByID[zID]

		if !ok {
			z = &Zone{}

			if ac.ZoneByID == nil {
				ac.ZoneByID = map[string]*Zone{}
			}

			ac.ZoneByID[zID] = z
		}

		z.TargetTemp = v
	}
}
