package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

func TestWorkerPoolUsage(t *testing.T) {
	t.Run("Pool not running with concurrency 0", WorkerPoolTest(0, false))
	t.Run("Pool not running with concurrency 1", WorkerPoolTest(1, false))
	t.Run("Pool running with concurrency 2", WorkerPoolTest(2, true))
	t.Run("Pool running with concurrency 3", WorkerPoolTest(3, true))
}

func WorkerPoolTest(maxConcurrency int, shouldBeActive bool) func(t *testing.T) {
	return func(t *testing.T) {
		arbiter := test.NewArbiter(t)
		s, run := PrepareServer(t,
			server.WithMaxConcurrency(maxConcurrency),
			server.WithOnWorkerStart(func() {
				arbiter.ItsAFactThat("SERVER_POOL_WORKER_STARTED")
			}),
			server.WithOnWorkerEnd(func() {
				arbiter.ItsAFactThat("SERVER_POOL_WORKER_ENDED")

			}))
		s.RegisterHandler(defaultMsg.Payload, nilHandler)
		run()
		defer s.Shutdown(defaultCtx)

		c, connect := PrepareClient(t,
			client.WithTargetServer(s.Addr()),
			client.WithMaxConcurrency(maxConcurrency),
			client.WithOnWorkerStart(func() {
				arbiter.ItsAFactThat("CLIENT_POOL_WORKER_STARTED")
			}),
			client.WithOnWorkerEnd(func() {
				arbiter.ItsAFactThat("CLIENT_POOL_WORKER_ENDED")
			}),
		)
		c.RegisterHandler(defaultMsg.Payload, nilHandler)
		connect()
		defer c.Close(defaultCtx)

		_, err := c.Send(defaultMsg)
		require.NoError(t, err)
		_, err = s.BroadCast(defaultCtx, defaultMsg)
		require.NoError(t, err)
		if shouldBeActive {
			arbiter.RequireHappenedInOrder(
				"SERVER_POOL_WORKER_STARTED",
				"SERVER_POOL_WORKER_ENDED",
				"CLIENT_POOL_WORKER_STARTED",
				"CLIENT_POOL_WORKER_ENDED",
			)
			return
		}
		arbiter.RequireNoEvents()
	}
}
