package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit open: temporarily unavailable")

type state int

const (
	closed state = iota
	open
	halfOpen
)

// Config controls circuit breaker thresholds.
type Config struct {
	// Threshold is the number of consecutive failures before opening the circuit.
	Threshold int
	// Cooldown is how long the circuit stays open before allowing a probe.
	Cooldown time.Duration
}

// Default returns sensible production defaults.
func Default() Config {
	return Config{Threshold: 5, Cooldown: 60 * time.Second}
}

// Breaker is a thread-safe circuit breaker. Use Do to wrap any fallible call.
type Breaker struct {
	mu       sync.Mutex
	st       state
	failures int
	openAt   time.Time
	cfg      Config
}

// New creates a Breaker with the given config.
func New(cfg Config) *Breaker {
	return &Breaker{cfg: cfg}
}

// Do executes fn, tracking success/failure to open or close the circuit.
// Returns ErrOpen immediately when the circuit is open.
func (b *Breaker) Do(fn func() error) error {
	if err := b.allow(); err != nil {
		return err
	}
	err := fn()
	b.record(err)
	return err
}

func Run[T any](b *Breaker, fn func() (T, error)) (T, error) {
	var zero T
	if err := b.allow(); err != nil {
		return zero, err
	}
	res, err := fn()
	b.record(err)
	return res, err
}

func (b *Breaker) allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.st {
	case open:
		if time.Since(b.openAt) >= b.cfg.Cooldown {
			b.st = halfOpen
			return nil
		}
		return ErrOpen
	default:
		return nil
	}
}

func (b *Breaker) record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		b.failures++
		if b.st == halfOpen || b.failures >= b.cfg.Threshold {
			b.st = open
			b.openAt = time.Now()
		}
		return
	}
	b.failures = 0
	b.st = closed
}
