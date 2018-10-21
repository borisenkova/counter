package count

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

func (p *WorkerPool) process(source *Source, stop <-chan struct{}) {
	if p.canSpawnWorkers() {
		p.spawnWorker()
	}

	p.sendWork(source, stop)
}

func (p *WorkerPool) sendWork(source *Source, stop <-chan struct{}) {
	select {
	case <-stop:
	case p.tasks <- source:
	}
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
	pool := &WorkerPool{
		maxPoolSize:   poolSize,
		newWorkerFunc: workerFunc,
		wg:            wg,
		tasks:         tasks,
	}

	return pool
}
