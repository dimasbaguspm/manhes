package concurrent

import "sync"

// Collect runs tasks concurrently and collects every result.
type Collect[T any] struct {
	mu      sync.Mutex
	results []T
	wg      sync.WaitGroup
}

func (c *Collect[T]) Go(fn func() T) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		r := fn()
		c.mu.Lock()
		c.results = append(c.results, r)
		c.mu.Unlock()
	}()
}

func (c *Collect[T]) Wait() []T {
	c.wg.Wait()
	return c.results
}
