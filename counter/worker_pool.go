package counter

import (
	"sync"
)

type WorkerPool struct {
	maxPoolSize     int
	numberOfWorkers int
	tasks           chan *Source
	newWorkerFunc   NewWorker
	wg              *sync.WaitGroup
}

type NewWorker func(wg *sync.WaitGroup, tasks <-chan *Source)

func (p *WorkerPool) process(source *Source) {
	if p.numberOfWorkers == 0 {
		p.spawnWorker()
	}

	if p.canSpawnWorkers() {
		if ok := p.trySendWork(source); ok {
			return
		}
		p.spawnWorker()
	}

	p.sendWork(source)
}

func (p *WorkerPool) trySendWork(source *Source) bool {
	select {
	case p.tasks <- source:
	default:
		return false
	}
	return true
}

func (p *WorkerPool) sendWork(source *Source) {
	p.tasks <- source
}

func (p *WorkerPool) canSpawnWorkers() bool {
	return p.numberOfWorkers < p.maxPoolSize
}

func (p *WorkerPool) spawnWorker() {
	p.wg.Add(1)
	go p.newWorkerFunc(p.wg, p.tasks)
	p.numberOfWorkers++
}

func newWorkerPool(tasks chan *Source, poolSize int, workerFunc NewWorker, wg *sync.WaitGroup) *WorkerPool {
	pool := WorkerPool{
		maxPoolSize:   poolSize,
		newWorkerFunc: workerFunc,
		wg:            wg,
		tasks:         tasks,
	}

	return &pool
}
