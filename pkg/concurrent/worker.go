package concurrent

import (
	"context"
	"sync"
)

type WorkerPool[J any] struct {
	workers int
	buffer  int
	jobChan chan J
	wg      sync.WaitGroup
}

func NewWorkerPool[J any](workers, buffer int) *WorkerPool[J] {
	return &WorkerPool[J]{
		workers: workers,
		buffer:  buffer,
		jobChan: make(chan J, buffer),
	}
}

func (p *WorkerPool[J]) Start(ctx context.Context, handler func(context.Context, J)) {
	p.wg.Add(p.workers)
	for range p.workers {
		go func() {
			defer p.wg.Done()
			for job := range p.jobChan {
				if ctx.Err() != nil {
					return
				}
				handler(ctx, job)
			}
		}()
	}
}

func (p *WorkerPool[J]) Submit(ctx context.Context, job J) bool {
	select {
	case p.jobChan <- job:
		return true
	case <-ctx.Done():
		return false
	default:
		return false
	}
}

func (p *WorkerPool[J]) SubmitChan() chan<- J {
	return p.jobChan
}

func (p *WorkerPool[J]) Wait() {
	close(p.jobChan)
	p.wg.Wait()
}
