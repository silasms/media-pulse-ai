package workerpool

import (
	"context"
	"sync"
)

type TaskHandler[T any] func(ctx context.Context, task T)

type Pool[T any] struct {
	workerCount int
	taskChan    chan T
	handler     TaskHandler[T]
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func New[T any](workerCount int, queueCapacity int, handler TaskHandler[T]) *Pool[T] {
	if workerCount <= 0 {
		workerCount = 1
	}
	if queueCapacity <= 0 {
		queueCapacity = 10
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Pool[T]{
		workerCount: workerCount,
		taskChan:    make(chan T, queueCapacity),
		handler:     handler,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (p *Pool[T]) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool[T]) worker() {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			p.handler(p.ctx, task)
		case <-p.ctx.Done():
			for {
				select {
				case task, ok := <-p.taskChan:
					if !ok {
						return
					}
					p.handler(p.ctx, task)
				default:
					return
				}
			}
		}
	}
}

func (p *Pool[T]) Submit(task T) bool {
	select {
	case p.taskChan <- task:
		return true
	default:
		return false
	}
}

func (p *Pool[T]) Stop() {
	p.cancel()
	close(p.taskChan)
	p.wg.Wait()
}
