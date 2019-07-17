package myplace

import "sort"

// System represents the entire system.
type System struct {
	Details struct {
		AppVersion       string `json:"myAppRev,omitempty"`
		NeedsUpdate      bool   `json:"needsUpdate,omitempty"`
		TouchScreenModel string `json:"tspModel,omitempty"`
		HasMyAir         bool   `json:"hasAircons,omitempty"`
		HasMyLights      bool   `json:"hasLights,omitempty"`
	} `json:"system,omitempty"`
	AirCons    []*AirCon          `json:"-"`
	AirConByID map[string]*AirCon `json:"aircons,omitempty"`
}

func (s *System) populate() {
	for id, ac := range s.AirConByID {
		ac.populate(id)
		s.AirCons = append(s.AirCons, ac)
	}

	sort.Slice(
		s.AirCons,
		func(i, j int) bool {
			return s.AirCons[i].ID < s.AirCons[j].ID
		},
	)
}
