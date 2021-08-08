package goomerang_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.eloylp.dev/kit/test"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/server"
)

const (
	serverAddr = "0.0.0.0:3000"
)

type Arbiter struct {
	t         *testing.T
	successes map[string][]success
	L         sync.RWMutex
	errors    []error
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

// todo make assertions eventually
func (a *Arbiter) AssertHappened(event string) *Arbiter {
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		_, ok := a.successes[event]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event)
	return a
}

func (a *Arbiter) AssertHappenedInOrder(event1, event2 string) *Arbiter {
	var e1, e2 []success
	var ok bool
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		e1, ok = a.successes[event1]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event1)
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		e2, ok = a.successes[event2]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event2)
	firstE1 := e1[0]
	firstE2 := e2[0]
	require.True(a.t, firstE1.time.UnixNano() < firstE2.time.UnixNano(), "event %s happened at %v, but event %s happened at %v", event1, firstE1, event2, firstE2)
	return a
}

func (a *Arbiter) AssertHappenedTimes(event string, expectedCount int) *Arbiter {
	var times []success
	var ok bool
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		times, ok = a.successes[event]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event)
	require.Lenf(a.t, times, expectedCount, "event %s expected to happen %v times. Got %v", event, expectedCount, times)
	return a
}

func (a *Arbiter) ErrorHappened(err error) {
	a.L.Lock()
	defer a.L.Unlock()
	a.errors = append(a.errors, err)
}

func (a *Arbiter) AssertNoErrors() {
	a.L.RLock()
	defer a.L.RUnlock()
	var msg strings.Builder
	for i, err := range a.errors {
		msg.WriteString(fmt.Sprintf("error %v: type %T: %v \n", i, err, err))
	}
	require.Len(a.t, a.errors, 0, msg)
}

func PrepareServer(t *testing.T, opts ...server.Option) *server.Server {
	t.Helper()
	opts = append(opts, server.WithListenAddr(serverAddr))
	s, err := server.NewServer(opts...)
	if err != nil {
		t.Fatal(err)
	}
	go s.Run()
	test.WaitTCPService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	return s
}

func PrepareClient(t *testing.T, opts ...client.Option) *client.Client {
	opts = append(opts, client.WithTargetServer(serverAddr))
	c, err := client.NewClient(opts...)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = c.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
