package count

import (
	"context"
	"log"
	"sync"
)

type workerPool struct {
	ctx             context.Context
	maxPoolSize     int
	numberOfWorkers int
	newWorkerFunc   newWorker
	wg              *sync.WaitGroup
	tasks           chan source
	results         chan *result
}

type newWorker func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan source, results chan<- *result)

func (p *workerPool) consume(tasks <-chan source) <-chan *result {
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

func (p *workerPool) waitForWorkersStop() {
	p.wg.Wait()
}

func (p *workerPool) process(source source) {
	if p.canSpawnWorkers() {
		p.spawnWorker()
	}

	p.sendWork(source)
}

func (p *workerPool) sendWork(source source) {
	select {
	case <-p.ctx.Done():
	case p.tasks <- source:
	}
}

func (p *workerPool) canSpawnWorkers() bool {
	return p.numberOfWorkers < p.maxPoolSize
}

func (p *workerPool) spawnWorker() {
	p.wg.Add(1)
	go p.newWorkerFunc(p.ctx, p.wg, p.tasks, p.results)
	p.numberOfWorkers++
}

func newWorkerPool(ctx context.Context, poolSize int, workerFunc newWorker) *workerPool {
	pool := &workerPool{
		ctx:           ctx,
		maxPoolSize:   poolSize,
		newWorkerFunc: workerFunc,
		wg:            &sync.WaitGroup{},
		tasks:         make(chan source),
		results:       make(chan *result),
	}

	return pool
}

func workerFunc(substring []byte) newWorker {
	return func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan source, results chan<- *result) {
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

				results <- &result{subtotal: subtotal, origin: source.origin(), error: err}
			case <-ctx.Done():
				return
			}
		}
	}
}
