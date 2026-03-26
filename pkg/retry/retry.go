package retry

import (
	"context"
	"time"
)

// Do retries fn up to maxAttempts times with exponential backoff starting at base.
func Do(ctx context.Context, maxAttempts int, base time.Duration, fn func() error) error {
	var err error
	wait := base
	for i := range maxAttempts {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			wait *= 2
		}
		if err = fn(); err == nil {
			return nil
		}
	}
	return err
}
