package goomerang

import (
	"errors"
)

type WorkerPool struct {
	ch chan struct{}
}

func (p *WorkerPool) Add() {
	p.ch <- struct{}{}
}

func (p *WorkerPool) Done() {
	<-p.ch
}

func NewWorkerPool(max int) (*WorkerPool, error) {
	if max < 1 {
		return nil, errors.New("workerPool: min concurrency should be 1")
	}
	return &WorkerPool{ch: make(chan struct{}, max)}, nil
}
