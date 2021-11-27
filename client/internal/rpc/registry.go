package rpc

import (
	"context"
	"fmt"
	"sync"
)

type Registry struct {
	r map[string]chan *MultiReply
	l *sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		r: map[string]chan *MultiReply{},
		l: &sync.Mutex{},
	}
}

func (r *Registry) CreateListener(id string) {
	r.l.Lock()
	defer r.l.Unlock()
	r.r[id] = make(chan *MultiReply, 1)
}

func (r *Registry) SubmitResult(id string, m *MultiReply) error {
	r.l.Lock()
	defer r.l.Unlock()
	ch, ok := r.r[id]
	if !ok {
		return fmt.Errorf("rpcregistry: cannot find key for %s", id)
	}
	ch <- m
	return nil
}

func (r *Registry) ResultFor(ctx context.Context, id string) (*MultiReply, error) {
	r.l.Lock()
	ch, ok := r.r[id]
	r.l.Unlock()
	if !ok {
		return nil, fmt.Errorf("rpcregistry: cannot find result for key for %s", id)
	}
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("rpcregistry: %w", ctx.Err())
	case reply := <-ch:
		delete(r.r, id)
		return reply, nil
	}
}