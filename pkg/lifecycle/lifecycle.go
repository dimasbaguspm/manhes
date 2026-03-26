package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithShutdown returns a context that is cancelled when the process receives
// SIGINT or SIGTERM. The caller must invoke the returned stop function.
func WithShutdown(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}
