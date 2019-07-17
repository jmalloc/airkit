package manager

import (
	"fmt"

	"github.com/brutella/hc/accessory"
	"github.com/jmalloc/airkit/myplace"
)

// NewBridge returns a new thermostat attached to the given system.
func NewBridge(s *myplace.System) *accessory.Bridge {
	return accessory.NewBridge(
		accessory.Info{
			Name:         "MyPlace",
			Manufacturer: "Advantage Air & James Harris",
			Model:        s.Details.TouchScreenModel,
			SerialNumber: "Unknown",
			FirmwareRevision: fmt.Sprintf(
				"MyPlace v%s & AirKit v%s",
				s.Details.AppVersion,
				"0.0.0",
			),
		},
	)
}
