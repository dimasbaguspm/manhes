package eventbus

import (
	"context"
	"log/slog"
	"sync"
)

// InMemBus is an in-memory implementation of Bus.
type InMemBus struct {
	handlers map[string][]Handler
	mu       sync.RWMutex
	log      *slog.Logger
}

// NewInMem creates a new in-memory event bus.
func NewInMem() *InMemBus {
	return &InMemBus{
		handlers: make(map[string][]Handler),
		log:      slog.With("component", "eventbus"),
	}
}

// Subscribe registers a handler for the given event name.
func (b *InMemBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	b.log.Debug("subscribed handler", "event", eventName, "total_handlers", len(b.handlers[eventName]))
}

// Publish delivers the event to all registered handlers asynchronously.
func (b *InMemBus) Publish(ctx context.Context, eventName string, event Event) error {
	b.mu.RLock()
	handlers, ok := b.handlers[eventName]
	b.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		return nil
	}

	for _, h := range handlers {
		go func(h Handler) {
			// Use background context so handlers aren't cancelled when caller's context is cancelled.
			if err := h(context.Background(), event); err != nil {
				b.log.Warn("handler returned error", "event", eventName, "err", err)
			}
		}(h)
	}
	return nil
}

// Close is a no-op for in-memory bus. Implements Bus interface.
func (b *InMemBus) Close() error {
	return nil
}
