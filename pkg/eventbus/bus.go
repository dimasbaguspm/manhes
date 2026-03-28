package eventbus

import "context"

// Event is the base interface for all domain events.
type Event interface{}

// Handler is the function signature for event subscribers.
type Handler func(ctx context.Context, event Event) error

// Bus is the event bus interface. Subscribers register handlers for specific
// event names, and publishers broadcast events to all registered handlers.
type Bus interface {
	Subscribe(eventName string, handler Handler)
	Publish(ctx context.Context, eventName string, event Event) error
	Close() error
}
