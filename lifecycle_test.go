package goomerang_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/internal/ws"
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

func statusChangesHook(a *test.Arbiter, side string) func(status uint32) {
	side = strings.ToUpper(side)
	return func(status uint32) {
		switch status {
		case ws.StatusNew:
			a.ItsAFactThat(side + "_WAS_NEW")
		case ws.StatusRunning:
			a.ItsAFactThat(side + "_WAS_RUNNING")
		case ws.StatusClosing:
			a.ItsAFactThat(side + "_WAS_CLOSING")
		case ws.StatusClosed:
			a.ItsAFactThat(side + "_WAS_CLOSED")
		}
	}
}

func noErrorHook(a *test.Arbiter) func(err error) {
	return func(err error) {
		a.ErrorHappened(err)
	}
}

func TestShutdownProcedureServerSideInit(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}),
		server.WithOnStatusChangeHook(statusChangesHook(arbiter, "server")),
		server.WithOnErrorHook(noErrorHook(arbiter)),
	)
	run()
	_, connect := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}),
		client.WithOnStatusChangeHook(statusChangesHook(arbiter, "client")),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	connect()
	require.NoError(t, s.Shutdown(defaultCtx))

	arbiter.RequireNoErrors()
	arbiter.RequireHappenedInOrder(
		"SERVER_WAS_NEW",
		"SERVER_WAS_RUNNING",
		"CLIENT_WAS_NEW",
		"CLIENT_WAS_RUNNING",
		"SERVER_WAS_CLOSING",
		"CLIENT_WAS_CLOSING",
		"CLIENT_WAS_CLOSED",
		"CLIENT_PROPERLY_CLOSED",
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

func TestClientCloseWhenServerClosed(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c, connect := PrepareClient(t)
	connect()
	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)

	err = c.Close(defaultCtx)
	require.NoError(t, err)
}

func TestServerSendsCloseToAllClients(t *testing.T) {
	arbiter := test.NewArbiter(t)

	errHook := func(err error) {
		arbiter.ErrorHappened(err)
	}
	s, run := PrepareServer(t, server.WithOnErrorHook(errHook))
	run()

	_, connect1 := PrepareClient(t, client.WithOnErrorHook(errHook), client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_CLOSE")
	}))
	connect1()

	_, connect2 := PrepareClient(t, client.WithOnErrorHook(errHook), client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_CLOSE")
	}))
	connect2()

	require.NoError(t, s.Shutdown(defaultCtx))

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_CLOSE")
	arbiter.RequireHappened("CLIENT2_RECEIVED_CLOSE")
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

func TestClientCannotSendMessagesIfClosed(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	c, connect := PrepareClient(t)
	connect()

	err := c.Close(defaultCtx)
	require.NoError(t, err)

	_, err = c.Send(defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "expected not to be able to send messages on closed status")
}

func TestClientCannotSendSyncMessagesIfClosed(t *testing.T) {

	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	c, connect := PrepareClient(t)
	connect()

	err := c.Close(defaultCtx)
	require.NoError(t, err)

	_, _, err = c.SendSync(defaultCtx, defaultMsg)
	assert.ErrorIs(t, err, client.ErrNotRunning, "expected not to be able to sendSync messages on closed status")
}

func TestServerCannotSendSyncMessagesIfClosed(t *testing.T) {
	s, run := PrepareServer(t)
	run()

	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)

	_, _, err = s.BroadCast(defaultCtx, defaultMsg)
	assert.EqualErrorf(t, err, "server: not running", "expected not to be able to sendSync messages on closed status")
}
