package goomerang_test

import (
	"sync"
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
