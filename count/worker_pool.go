package count

import (
	"context"
	"log"
	"sync"
)

type WorkerPool struct {
	ctx             context.Context
	maxPoolSize     int
	numberOfWorkers int
	newWorkerFunc   NewWorker
	wg              *sync.WaitGroup
	tasks           chan *Source
	results         chan *Result
}

type NewWorker func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan *Source, results chan<- *Result)

func (p *WorkerPool) consume(tasks <-chan *Source) <-chan *Result {
	p.wg.Add(1)
	go func() {
		defer func() {
			close(p.tasks)
			p.wg.Done()
			p.wg.Wait()
			close(p.results)
		}()

		for {
			select {
			case <-p.ctx.Done():
				return
			case source, hasMore := <-tasks:
				if !hasMore {
					return
				}

				p.process(source)
			}
		}
	}()

	return p.results
}

func (p *WorkerPool) waitForWorkersStop() {
	p.wg.Wait()
}

func (p *WorkerPool) process(source *Source) {
	if p.canSpawnWorkers() {
		p.spawnWorker()
	}

	p.sendWork(source)
}

func (p *WorkerPool) sendWork(source *Source) {
	select {
	case <-p.ctx.Done():
	case p.tasks <- source:
	}
}

func (p *WorkerPool) canSpawnWorkers() bool {
	return p.numberOfWorkers < p.maxPoolSize
}

func (p *WorkerPool) spawnWorker() {
	p.wg.Add(1)
	go p.newWorkerFunc(p.ctx, p.wg, p.tasks, p.results)
	p.numberOfWorkers++
}

func newWorkerPool(ctx context.Context, poolSize int, workerFunc NewWorker) *WorkerPool {
	pool := &WorkerPool{
		ctx:           ctx,
		maxPoolSize:   poolSize,
		newWorkerFunc: workerFunc,
		wg:            &sync.WaitGroup{},
		tasks:         make(chan *Source),
		results:       make(chan *Result),
	}

	return pool
}

func workerFunc(substring []byte) NewWorker {
	return func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan *Source, results chan<- *Result) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in workerFunc", r)
			}
		}()
		defer wg.Done()

		buf := make([]byte, averageWebpageSize)
		for {
			select {
			case source, hasMore := <-tasks:
				if !hasMore {
					return
				}

				subtotal, err := processSource(ctx, source, buf, substring)
				if ctx.Err() != nil {
					return
				}

				results <- &Result{subtotal: subtotal, origin: source.origin, error: err}
			case <-ctx.Done():
				return
			}
		}
	}
}
