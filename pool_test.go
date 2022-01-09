package goomerang_test

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/goomerang"
)

func TestWorkerPool(t *testing.T) {
	var concurrent int32
	c, err := goomerang.NewWorkerPool(10)
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
		fmt.Println(MaxConcurrent)
		return int32(10) == MaxConcurrent
	}, time.Second, time.Millisecond)
}

func TestNewWorkerPool(t *testing.T) {
	_, err := goomerang.NewWorkerPool(0)
	assert.EqualError(t, err, "workerPool: min concurrency should be 1")
}
