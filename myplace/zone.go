package myplace

import "fmt"

// ZoneState is an enumeration of the states of a zone.
type ZoneState string

const (
	// ZoneStateOpen means the damper for this zone is open.
	ZoneStateOpen ZoneState = "open"

	// ZoneStateClosed means the damper for this zone is closed.
	ZoneStateClosed ZoneState = "close" // note: value is 'close', without the trailing 'd'.
)

func (s ZoneState) String() string {
	switch s {
	case ZoneStateOpen:
		return "on"
	case ZoneStateClosed:
		return "off"
	default:
		return "unknown"
	}
}

// ZoneError is an enumeration of the zone error codes.
type ZoneError int

const (
	// ZoneErrorNone is an error code indicating that there are no known errors.
	ZoneErrorNone ZoneError = 0

	// ZoneErrorNoSignal is the error code given when a zone's temperature
	// sensor can not be reached (due to signal loss, or a dead battery, for
	// example). Referred to as AA84 in the MyPlace application.
	ZoneErrorNoSignal ZoneError = 2
)

func (e ZoneError) String() string {
	switch e {
	case ZoneErrorNone:
		return "ok"
	case ZoneErrorNoSignal:
		return "no signal from temp. sensor"
	default:
		return "unknown error"
	}
}

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
	Error            ZoneError `json:"error,omitempty"`
}

func (z *Zone) populate(id string) {
	z.ID = id
}

// SetMyZone returns a command that sets "MyZone" of an air-conditioning unit.
func SetMyZone(id string, zone *Zone) Command {
	return Command{
		desc: fmt.Sprintf("set the %s MyZone to #%d (%s)", id, zone.Number, zone.Name),
		apply: func(req map[string]*AirCon) {
			ac, ok := req[id]

			if !ok {
				ac = &AirCon{}
				req[id] = ac
			}

			ac.Details.MyZoneNumber = zone.Number
		},
	}
}

// SetZoneState returns a command that sets state of a zone.
func SetZoneState(id string, zone *Zone, v ZoneState) Command {
	return Command{
		desc: fmt.Sprintf("set %s#%d (%s) to %s", id, zone.Number, zone.Name, v),
		apply: func(req map[string]*AirCon) {
			ac, ok := req[id]

			if !ok {
				ac = &AirCon{}
				req[id] = ac
			}

			z, ok := ac.ZoneByID[zone.ID]

			if !ok {
				z = &Zone{}

				if ac.ZoneByID == nil {
					ac.ZoneByID = map[string]*Zone{}
				}

				ac.ZoneByID[zone.ID] = z
			}

			z.State = v
		},
	}
}

// SetZoneTargetTemp returns a command that sets the fan mode of an
// air-conditioning unit.
func SetZoneTargetTemp(id string, zone *Zone, v float64) Command {
	return Command{
		desc: fmt.Sprintf("set %s#%d (%s) target temperature to %.1f°C", id, zone.Number, zone.Name, v),
		apply: func(req map[string]*AirCon) {
			ac, ok := req[id]

			if !ok {
				ac = &AirCon{}
				req[id] = ac
			}

			z, ok := ac.ZoneByID[zone.ID]

			if !ok {
				z = &Zone{}

				if ac.ZoneByID == nil {
					ac.ZoneByID = map[string]*Zone{}
				}

				ac.ZoneByID[zone.ID] = z
			}

			z.TargetTemp = v
		},
	}
}
