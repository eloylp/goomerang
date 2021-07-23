package goomerang_test

import (
	"context"
	"sync"

	"google.golang.org/protobuf/proto"
)

const (
	serverReceivedPing = "SERVER_RECEIVED_PING"
	clientReceivedPong = "CLIENT_RECEIVED_PONG"
)

type Arbiter struct {
	successes map[string]struct{}
	L         sync.RWMutex
}

func NewArbiter() *Arbiter {
	return &Arbiter{
		successes: make(map[string]struct{}),
	}
}

func (a *Arbiter) ItsAFactThat(event string) {
	a.L.Lock()
	defer a.L.Unlock()
	a.successes[event] = struct{}{}
}

func (a *Arbiter) AssertHappened(event string) bool {
	a.L.RLock()
	defer a.L.RUnlock()
	_, ok := a.successes[event]
	return ok
}

type FakeServerOpts struct {
}

func (f *FakeServerOpts) Send(ctx context.Context, msg proto.Message) error {
	panic("implement me")
}

func (f *FakeServerOpts) Shutdown(ctx context.Context) error {
	panic("implement me")
}
