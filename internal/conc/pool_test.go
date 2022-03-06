package conc_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/conc"
)

func TestWorkerPool(t *testing.T) {
	var concurrent int32
	c, err := conc.NewWorkerPool(10)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		c.Add()
		atomic.AddInt32(&concurrent, 1)
		go func() {
			time.Sleep(50 * time.Millisecond)
			c.Done()
			atomic.AddInt32(&concurrent, -1)
		}()
	}
	assert.Eventually(t, func() bool {
		MaxConcurrent := atomic.LoadInt32(&concurrent)
		return int32(10) == MaxConcurrent
	}, time.Second, time.Millisecond)
}

func TestNewWorkerPool(t *testing.T) {
	_, err := conc.NewWorkerPool(0)
	assert.EqualError(t, err, "workerPool: min concurrency should be 1")
}
