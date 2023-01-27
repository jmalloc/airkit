package manager

import "github.com/jmalloc/airkit/myplace"

const (
	acFanSpeedOverrideID = 1
)

const (
	zoneThermostatID = 1
)

func makeAirConAccessoryID(ac *myplace.AirCon, id uint32) uint64 {
	idMask := uint64(ac.Number) << 56
	return idMask | uint64(id)
}

func makeZoneAccessoryID(ac *myplace.AirCon, z *myplace.Zone, id uint32) uint64 {
	idMask := uint64(ac.Number)<<56 | uint64(z.Number)<<48
	return idMask | uint64(id)
}
