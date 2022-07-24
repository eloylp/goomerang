package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
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
		s.Handle(defaultMsg.Payload, nilHandler)
		run()
		defer s.Shutdown(defaultCtx)

		c, connect := PrepareClient(t,
			client.WithServerAddr(s.Addr()),
			client.WithMaxConcurrency(maxConcurrency),
			client.WithOnWorkerStart(func() {
				arbiter.ItsAFactThat("CLIENT_POOL_WORKER_STARTED")
			}),
			client.WithOnWorkerEnd(func() {
				arbiter.ItsAFactThat("CLIENT_POOL_WORKER_ENDED")
			}),
		)
		c.Handle(defaultMsg.Payload, nilHandler)
		connect()
		defer c.Close(defaultCtx)

		_, err := c.Send(defaultMsg)
		require.NoError(t, err)
		_, err = s.Broadcast(defaultCtx, defaultMsg)
		require.NoError(t, err)
		if shouldBeActive {
			arbiter.RequireHappenedInOrder(
				"SERVER_POOL_WORKER_STARTED",
				"SERVER_POOL_WORKER_ENDED",
			)
			arbiter.RequireHappenedInOrder(
				"CLIENT_POOL_WORKER_STARTED",
				"CLIENT_POOL_WORKER_ENDED",
			)
			return
		}
		arbiter.RequireNoEvents()
	}
}

func TestSendVariousMessagesWithNoConcurrency(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithMaxConcurrency(0))
	defer s.Shutdown(defaultCtx)

	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("SERVER_RECEIVED_MSG")
		_, _ = s.Send(msg)
	}))
	run()
	c, connect := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithMaxConcurrency(0),
	)
	defer c.Close(defaultCtx)

	c.Handle(defaultMsg.Payload, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT_RECEIVED_MSG")
	}))
	connect()

	_, err := c.Send(defaultMsg)
	require.NoError(t, err)
	_, err = c.Send(defaultMsg)
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappenedTimes("CLIENT_RECEIVED_MSG", 2)
	arbiter.RequireHappenedTimes("SERVER_RECEIVED_MSG", 2)
}
