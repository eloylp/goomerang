package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureServerSideInit(t *testing.T) {
	serverArbiter := test.NewArbiter(t)
	clientArbiter := test.NewArbiter(t)

	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		serverArbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}),
		server.WithOnStatusChangeHook(statusChangesHook(serverArbiter, "server")),
		server.WithOnErrorHook(noErrorHook(serverArbiter)),
	)
	run()
	_, connect := PrepareClient(t, client.WithOnCloseHook(func() {
		clientArbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}),
		client.WithOnStatusChangeHook(statusChangesHook(clientArbiter, "client")),
		client.WithOnErrorHook(noErrorHook(clientArbiter)),
	)
	connect()
	require.NoError(t, s.Shutdown(defaultCtx))

	serverArbiter.RequireNoErrors()
	clientArbiter.RequireNoErrors()

	serverArbiter.RequireHappenedInOrder(
		"SERVER_WAS_NEW",
		"SERVER_WAS_RUNNING",
		"SERVER_WAS_CLOSING",
		"SERVER_WAS_CLOSED",
		"SERVER_PROPERLY_CLOSED",
	)

	clientArbiter.RequireHappenedInOrder(
		"CLIENT_WAS_NEW",
		"CLIENT_WAS_RUNNING",
		"CLIENT_WAS_CLOSING",
		"CLIENT_WAS_CLOSED",
		"CLIENT_PROPERLY_CLOSED",
	)
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

func TestServerCannotSendMessagesIfNotRunning(t *testing.T) {
	s, run := PrepareServer(t)

	_, _, err := s.BroadCast(defaultCtx, defaultMsg)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not broadcast messages if not connected")

	run()

	err = s.Shutdown(defaultCtx)
	require.NoError(t, err)

	_, _, err = s.BroadCast(defaultCtx, defaultMsg)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not broadcast messages on closed status")
}

func TestServerCannotShutdownIfNotRunning(t *testing.T) {
	s, _ := PrepareServer(t)
	err := s.Shutdown(defaultCtx)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not shutdown if not running")
}

func TestServerCannotShutdownIfClosed(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(defaultCtx)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not shutdown if already closed")
}

func TestServerCannotRunIfAlreadyRunning(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	defer s.Shutdown(defaultCtx)
	err := s.Run()
	assert.ErrorIs(t, err, server.ErrAlreadyRunning, "should not run if already running")
}
