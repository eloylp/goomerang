package test

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Arbiter struct {
	t        *testing.T
	evCount  map[string]int
	evSeries []string
	L        sync.RWMutex
	errors   []error
}

func NewArbiter(t *testing.T) *Arbiter {
	return &Arbiter{
		t:        t,
		evCount:  make(map[string]int),
		evSeries: []string{},
	}
}

func (a *Arbiter) ItsAFactThat(event string) {
	a.L.Lock()
	defer a.L.Unlock()
	a.evCount[event]++
	a.evSeries = append(a.evSeries, event)
}

func (a *Arbiter) RequireHappened(event string) *Arbiter {
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		_, ok := a.evCount[event]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event)
	return a
}

func (a *Arbiter) RequireHappenedInOrder(events ...string) *Arbiter {
	assert.Eventually(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		return reflect.DeepEqual(a.evSeries, events)
	}, time.Second, time.Millisecond)
	require.Equal(a.t, events, a.evSeries, "expected event series doesnt match")
	return a
}

func (a *Arbiter) RequireHappenedTimes(event string, expectedCount int) *Arbiter {
	var count int
	var ok bool
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		count, ok = a.evCount[event]
		if !ok {
			return false
		}
		return count == expectedCount
	}, time.Second, time.Millisecond, "event %s not happened", event)
	require.Lenf(a.t, count, expectedCount, "event %s expected to happen %v times. Got %v", event, expectedCount, count)
	return a
}

func (a *Arbiter) ErrorHappened(err error) {
	a.L.Lock()
	defer a.L.Unlock()
	a.errors = append(a.errors, err)
}

func (a *Arbiter) RequireNoErrors() {
	a.L.RLock()
	defer a.L.RUnlock()
	var msg strings.Builder
	for i, err := range a.errors {
		msg.WriteString(fmt.Sprintf("error %v: type %T: %v \n", i, err, err))
	}
	require.Len(a.t, a.errors, 0, msg)
}

func (a *Arbiter) RequireError(errMsg string) {
	require.Eventuallyf(a.t, func() bool {
		a.L.RLock()
		defer a.L.RUnlock()
		for _, err := range a.errors {
			if strings.Contains(err.Error(), errMsg) {
				return true
			}
		}
		return false
	}, time.Second, time.Millisecond, "expected error %q not happened in expected time.", errMsg)
	a.L.RLock()
	defer a.L.RUnlock()
}
