package conc

import (
	"errors"
	"sync"
)

type WorkerPool struct {
	ch chan struct{}
	wg *sync.WaitGroup
}

func (p *WorkerPool) Add() {
	p.ch <- struct{}{}
	p.wg.Add(1)
}

func (p *WorkerPool) Done() {
	<-p.ch
	p.wg.Done()
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

func NewWorkerPool(max int) (*WorkerPool, error) {
	if max < 1 {
		return nil, errors.New("workerPool: min concurrency should be 1")
	}
	return &WorkerPool{
		ch: make(chan struct{}, max),
		wg: &sync.WaitGroup{},
	}, nil
}
