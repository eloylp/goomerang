package test

import (
	"errors"
	"fmt"
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
	l        sync.RWMutex
	errors   []error
}

func NewArbiter(t *testing.T) *Arbiter {
	return &Arbiter{
		t:        t,
		evCount:  make(map[string]int),
		evSeries: []string{},
	}
}

func (a *Arbiter) EvCount() map[string]int {
	a.l.Lock()
	defer a.l.Unlock()
	snapshot := make(map[string]int, len(a.evCount))
	for k, v := range a.evCount {
		snapshot[k] = v
	}
	return snapshot
}

func (a *Arbiter) Errors() []error {
	a.l.Lock()
	defer a.l.Unlock()
	snapshot := make([]error, len(a.errors))
	for i := 0; i < len(a.errors); i++ {
		snapshot[i] = a.errors[i]
	}
	return a.errors
}

func (a *Arbiter) ItsAFactThat(event string) {
	a.l.Lock()
	defer a.l.Unlock()
	a.evCount[event]++
	a.evSeries = append(a.evSeries, event)
}

func (a *Arbiter) RequireHappened(event string) *Arbiter {
	require.Eventuallyf(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()
		_, ok := a.evCount[event]
		return ok
	}, time.Second, time.Millisecond, "event %s not happened", event)
	return a
}

func (a *Arbiter) RequireHappenedInOrder(events ...string) *Arbiter {
	passed := assert.Eventually(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()

		if len(a.evSeries) < len(events) {
			return false
		}
		// Iterate over all events in the store,
		// searching the expected chain of events is
		// contained in the happened one. It allows
		// other events interleaving.
		var current int
		for i := 0; i < len(a.evSeries); i++ {
			if len(events)-1 == current {
				break
			}
			if a.evSeries[i] == events[current] {
				current++
			}
		}
		return current == len(events)-1
	}, time.Second, time.Millisecond)
	if !passed {
		a.t.Logf("expected event series: %v, its not contained in the event store: %v", events, a.evSeries)
	}
	return a
}

func (a *Arbiter) RequireHappenedTimes(event string, expectedCount int) *Arbiter {
	var count int
	var ok bool
	assert.Eventuallyf(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()
		count, ok = a.evCount[event]
		if !ok {
			return false
		}
		return count == expectedCount
	}, time.Second, time.Millisecond, "event %s not happened", event)
	require.Equal(a.t, count, expectedCount, "event %s expected to happen %v times. Got %v", event, expectedCount, count)
	return a
}

func (a *Arbiter) ErrorHappened(err error) {
	if err != nil {
		a.l.Lock()
		defer a.l.Unlock()
		a.errors = append(a.errors, err)
	}
}

func (a *Arbiter) RequireNoErrors() {
	a.l.RLock()
	defer a.l.RUnlock()
	var msg strings.Builder
	for i, err := range a.errors {
		msg.WriteString(fmt.Sprintf("error %v: type %T: %v \n", i, err, err))
	}
	require.Len(a.t, a.errors, 0, msg)
}

func (a *Arbiter) RequireError(errMsg string) {
	require.Eventuallyf(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()
		for _, err := range a.errors {
			if strings.Contains(err.Error(), errMsg) {
				return true
			}
		}
		return false
	}, time.Second, time.Millisecond, "expected error %q not happened in expected time.", errMsg)
	a.l.RLock()
	defer a.l.RUnlock()
}

func (a *Arbiter) RequireErrorIs(err error) {
	require.Eventuallyf(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()
		for i := range a.errors {
			if errors.Is(a.errors[i], err) {
				return true
			}
		}
		return false
	}, time.Second, time.Millisecond, "expected error: %q : not found in chain, found: %q", err, a.errors)
}

func (a *Arbiter) RequireNoEvents() {
	require.Eventuallyf(a.t, func() bool {
		a.l.RLock()
		defer a.l.RUnlock()
		return len(a.evSeries) == 0
	}, time.Second, time.Millisecond, "no events expected, but got %v", len(a.evSeries))
}
