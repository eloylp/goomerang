package conc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/goomerang/internal/conc"
)

func TestNewWorkerPool(t *testing.T) {
	_, err := conc.NewWorkerPool(-1)
	assert.EqualError(t, err, "workerPool: min concurrency should be 0")
}
