package count

import (
	"context"
	"log"
	"sync"
	"time"
)

type workerPool struct {
	ctx             context.Context
	maxPoolSize     int
	numberOfWorkers int
	workersMx       sync.Mutex
	newWorker       newWorker
	wg              *sync.WaitGroup
	tasks           chan source
	results         chan *result
}

type newWorker func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan source, results chan<- *result, stopped chan struct{})

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
	p.spawnWorkerIfPossible()
	p.sendWork(source)
}

func (p *workerPool) sendWork(source source) {
	select {
	case <-p.ctx.Done():
	case p.tasks <- source:
	}
}

func (p *workerPool) spawnWorkerIfPossible() {
	p.workersMx.Lock()
	defer p.workersMx.Unlock()

	if p.canSpawnWorkers() {
		p.incNumberOfWorkers()

		closedOnWorkerStop := p.spawnWorker()
		go func() {
			<-closedOnWorkerStop
			p.decNumberOfWorkers()
		}()
	}
}

func (p *workerPool) canSpawnWorkers() bool {
	return p.numberOfWorkers < p.maxPoolSize
}

func (p *workerPool) spawnWorker() <-chan struct{} {
	p.wg.Add(1)
	stopped := make(chan struct{})
	go p.newWorker(p.ctx, p.wg, p.tasks, p.results, stopped)
	return stopped
}

// incNumberOfWorkers increments current number of running workers. Warning: it
// supposed that workersMx is locked by the caller of the method
func (p *workerPool) incNumberOfWorkers() {
	p.numberOfWorkers++
}

func (p *workerPool) decNumberOfWorkers() {
	p.workersMx.Lock()
	defer p.workersMx.Unlock()
	p.numberOfWorkers--
}

func (p *workerPool) getNumberOfWorkers() int {
	p.workersMx.Lock()
	defer p.workersMx.Unlock()
	return p.numberOfWorkers
}

func newWorkerPool(ctx context.Context, poolSize int, workerFunc newWorker) *workerPool {
	pool := &workerPool{
		ctx:         ctx,
		maxPoolSize: poolSize,
		newWorker:   workerFunc,
		wg:          &sync.WaitGroup{},
		tasks:       make(chan source),
		results:     make(chan *result),
	}

	return pool
}

func workerFunc(substring []byte, workerShutdownTimeout time.Duration) newWorker {
	return func(ctx context.Context, wg *sync.WaitGroup, tasks <-chan source, results chan<- *result, stopped chan struct{}) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered in workerFunc", r)
			}
		}()
		defer func() {
			close(stopped)
			wg.Done()
		}()

		buf := make([]byte, averageWebpageSize)
		t := time.NewTimer(workerShutdownTimeout)
		for {
			select {
			case source, hasMore := <-tasks:
				if !hasMore {
					return
				}

				subtotal, err := processSource(ctx, source, buf, substring)
				if ctx.Err() != nil && ctx.Err() == context.Canceled {
					return
				}

				results <- &result{subtotal: subtotal, origin: source.origin(), error: err}

				if !t.Stop() {
					<-t.C
				}
				t.Reset(workerShutdownTimeout)
			case <-ctx.Done():
				return
			case <-t.C:
				return
			}
		}
	}
}
