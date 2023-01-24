package manager

import (
	"fmt"

	"github.com/brutella/hap/accessory"
	"github.com/jmalloc/airkit/myplace"
)

// NewBridge returns a new thermostat attached to the given system.
func NewBridge(
	version string,
	s *myplace.System,
) *accessory.Bridge {
	return accessory.NewBridge(
		accessory.Info{
			Name:         "MyPlace",
			Manufacturer: "Advantage Air & James Harris",
			Model:        s.Details.TouchScreenModel,
			SerialNumber: "Unknown",
			Firmware: fmt.Sprintf(
				"MyPlace v%s & AirKit v%s",
				s.Details.AppVersion,
				version,
			),
		},
	)
}
