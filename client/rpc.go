package client

import (
	"context"
	"fmt"
	"sync"

	"go.eloylp.dev/goomerang/message"
)

type registry struct {
	r map[string]chan *message.Message
	l *sync.Mutex
}

func newRegistry() *registry {
	return &registry{
		r: map[string]chan *message.Message{},
		l: &sync.Mutex{},
	}
}

func (r *registry) createListener(id string) {
	r.l.Lock()
	defer r.l.Unlock()
	r.r[id] = make(chan *message.Message, 1)
}

func (r *registry) submitResult(id string, m *message.Message) error {
	r.l.Lock()
	defer r.l.Unlock()
	ch, ok := r.r[id]
	if !ok {
		return fmt.Errorf("rpcregistry: cannot find key for %s", id)
	}
	ch <- m
	return nil
}

func (r *registry) resultFor(ctx context.Context, id string) (*message.Message, error) {
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
