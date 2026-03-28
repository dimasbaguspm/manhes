package eventbus

import (
	"manga-engine/pkg/eventbus"
)

// New returns a generic in-memory event bus.
func New() eventbus.Bus {
	return eventbus.NewInMem()
}
