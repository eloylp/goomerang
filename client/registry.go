package client

import (
	"context"
	"fmt"
	"sync"

	"go.eloylp.dev/goomerang/message"
)

type requestRegistry struct {
	r map[string]chan *message.Message
	l *sync.Mutex
}

func newRegistry() *requestRegistry {
	return &requestRegistry{
		r: map[string]chan *message.Message{},
		l: &sync.Mutex{},
	}
}

func (r *requestRegistry) createListener(id string) {
	r.l.Lock()
	defer r.l.Unlock()
	r.r[id] = make(chan *message.Message, 1)
}

func (r *requestRegistry) submitResult(id string, m *message.Message) error {
	r.l.Lock()
	defer r.l.Unlock()
	ch, ok := r.r[id]
	if !ok {
		return fmt.Errorf("request registry: cannot find key for %s", id)
	}
	ch <- m
	return nil
}

func (r *requestRegistry) resultFor(ctx context.Context, id string) (*message.Message, error) {
	r.l.Lock()
	ch, ok := r.r[id]
	r.l.Unlock()
	if !ok {
		return nil, fmt.Errorf("request registry: cannot find result for key for %s", id)
	}
	select {
	case <-ctx.Done():
		r.l.Lock()
		delete(r.r, id)
		r.l.Unlock()
		return nil, fmt.Errorf("request registry: %w", ctx.Err())
	case reply := <-ch:
		r.l.Lock()
		delete(r.r, id)
		r.l.Unlock()
		return reply, nil
	}
}
