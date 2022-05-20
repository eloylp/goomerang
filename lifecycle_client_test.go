package goomerang_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureClientSideInit(t *testing.T) {
	arbiter := test.NewArbiter(t)

	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}),
		server.WithOnStatusChangeHook(statusChangesHook(arbiter, "server")),
		server.WithOnErrorHook(noErrorHook(arbiter)),
	)
	run()

	c, connect := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}),
		client.WithOnStatusChangeHook(statusChangesHook(arbiter, "client")),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	connect()

	err := c.Close(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(context.Background())
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappenedInOrder(
		"SERVER_WAS_NEW",
		"SERVER_WAS_RUNNING",
		"CLIENT_WAS_NEW",
		"CLIENT_WAS_RUNNING",
		"CLIENT_WAS_CLOSING",
		"CLIENT_WAS_CLOSED",
		"CLIENT_PROPERLY_CLOSED",
		"SERVER_WAS_CLOSING",
		"SERVER_WAS_CLOSED",
		"SERVER_PROPERLY_CLOSED",
	)
}

func TestClientNormalClose(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()
	c, connect := PrepareClient(t)
	connect()
	require.NoError(t, c.Close(defaultCtx))
}

func TestClientShouldBeClosedWhenServerCloses(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c, connect := PrepareClient(t)
	connect()
	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)

	err = c.Close(defaultCtx)
	require.ErrorIs(t, err, client.ErrNotRunning)
}

func TestClientCanBeResumed(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	c, connect := PrepareClient(t)
	connect()

	require.NoError(t, c.Close(defaultCtx))
	require.NoError(t, c.Connect(defaultCtx))
	require.NoError(t, c.Close(defaultCtx))
}

func TestClientCannotConnectForSecondTime(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	c, connect := PrepareClient(t)
	connect()
	defer c.Close(defaultCtx)

	assert.EqualError(t, c.Connect(defaultCtx), "client: already connected")
}

func TestClientCannotSendMessagesIfNotRunning(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	// Cannot send if not connected
	c, connect := PrepareClient(t)
	_, err := c.Send(defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "should not send messages if not connected")

	_, _, err = c.SendSync(defaultCtx, defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "should not send sync messages if not connected")

	connect()

	err = c.Close(defaultCtx)
	require.NoError(t, err)

	// Cannot send if closed
	_, err = c.Send(defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "should not send messages on closed status")

	_, _, err = c.SendSync(defaultCtx, defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "should not send sync messages on closed status")
}

func TestClientCannotShutdownIfNotRunning(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	defer s.Shutdown(defaultCtx)

	c, connect := PrepareClient(t)
	assert.ErrorIs(t, c.Close(defaultCtx), client.ErrNotRunning, "should not shutdown if not running")
	connect()
	require.NoError(t, c.Close(defaultCtx))

	assert.ErrorIs(t, c.Close(defaultCtx), client.ErrNotRunning, "should not shutdown if already closed")
}
