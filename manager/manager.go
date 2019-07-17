package manager

import (
	"github.com/brutella/hc/accessory"
	"github.com/jmalloc/airkit/myplace"
)

// AccessoryManager is an interface for managing synchronization of state
// between homekit accessories and the MyPlace system.
type AccessoryManager interface {
	Accessories() []*accessory.Accessory
	Update(sys *myplace.System)
}
