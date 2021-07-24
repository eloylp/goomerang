package goomerang_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	serverReceivedPing = "SERVER_RECEIVED_PING"
	clientReceivedPong = "CLIENT_RECEIVED_PONG"
)

type Arbiter struct {
	t         *testing.T
	successes map[string][]success
	L         sync.RWMutex
}

type success struct {
	time time.Time
}

func NewArbiter(t *testing.T) *Arbiter {
	return &Arbiter{
		t:         t,
		successes: make(map[string][]success),
	}
}

func (a *Arbiter) ItsAFactThat(event string) {
	a.L.Lock()
	defer a.L.Unlock()
	a.successes[event] = append(a.successes[event], success{time: time.Now()})
}

func (a *Arbiter) AssertHappened(event string) *Arbiter {
	a.L.RLock()
	defer a.L.RUnlock()
	_, ok := a.successes[event]
	require.Truef(a.t, ok, "event %s not happened", event)
	return a
}

func (a *Arbiter) AssertHappenedInOrder(event1, event2 string) *Arbiter {
	a.L.RLock()
	defer a.L.RUnlock()
	e1, ok := a.successes[event1]
	require.Truef(a.t, ok, "event %s not happened", event1)
	e2, ok := a.successes[event2]
	require.Truef(a.t, ok, "event %s not happened", event1)

	firstE1 := e1[0]
	firstE2 := e2[0]

	require.True(a.t, firstE1.time.UnixNano() < firstE2.time.UnixNano(), "event %s happened at %v, but event %s happened at %v", event1, firstE1, event2, firstE2)

	return a
}

func (a *Arbiter) AssertHappenedTimes(event string, expectedCount int) *Arbiter {
	a.L.RLock()
	defer a.L.RUnlock()
	times, ok := a.successes[event]
	require.Truef(a.t, ok, "event %s not happened", event)
	require.Equalf(a.t, expectedCount, times, "event %s expected to happen %v times. Got %v", event, expectedCount, times)
	return a
}
