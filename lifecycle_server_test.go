//go:build integration

package goomerang_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureServerSideInit(t *testing.T) {
	t.Parallel()

	serverArbiter := test.NewArbiter(t)
	clientArbiter := test.NewArbiter(t)

	s, run := Server(t, server.WithOnCloseHook(func() {
		serverArbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}),
		server.WithOnStatusChangeHook(statusChangesHook(serverArbiter, "server")),
		server.WithOnErrorHook(noErrorHook(serverArbiter)),
	)
	run()
	_, connect := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnCloseHook(func() {
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
	t.Parallel()

	arbiter := test.NewArbiter(t)

	errHook := func(err error) {
		arbiter.ErrorHappened(err)
	}
	s, run := Server(t, server.WithOnErrorHook(errHook))
	run()

	_, connect1 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(errHook),
		client.WithOnCloseHook(func() {
			arbiter.ItsAFactThat("CLIENT1_RECEIVED_CLOSE")
		}))
	connect1()

	_, connect2 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(errHook),
		client.WithOnCloseHook(func() {
			arbiter.ItsAFactThat("CLIENT2_RECEIVED_CLOSE")
		}))
	connect2()

	require.NoError(t, s.Shutdown(defaultCtx))

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_CLOSE")
	arbiter.RequireHappened("CLIENT2_RECEIVED_CLOSE")
}

func TestServerCannotSendMessagesIfNotRunning(t *testing.T) {
	t.Parallel()

	s, run := Server(t)

	_, err := s.Broadcast(defaultCtx, defaultMsg())
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not broadcast messages if not connected")

	run()

	err = s.Shutdown(defaultCtx)
	require.NoError(t, err)

	_, err = s.Broadcast(defaultCtx, defaultMsg())
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not broadcast messages on closed status")
}

func TestServerCannotShutdownIfNotRunning(t *testing.T) {
	t.Parallel()

	s, _ := Server(t)
	err := s.Shutdown(defaultCtx)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not shutdown if not running")
}

func TestServerCannotShutdownIfClosed(t *testing.T) {
	t.Parallel()

	s, run := Server(t)
	run()
	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(defaultCtx)
	assert.ErrorIs(t, err, server.ErrNotRunning, "should not shutdown if already closed")
}

func TestServerCannotRunIfAlreadyRunning(t *testing.T) {
	t.Parallel()

	s, run := Server(t)
	run()
	defer s.Shutdown(defaultCtx)
	err := s.Run()
	assert.ErrorIs(t, err, server.ErrAlreadyRunning, "should not run if already running")
}

func TestServerCannotRunIfAlreadyClosed(t *testing.T) {
	t.Parallel()

	s, run := Server(t)
	run()
	require.NoError(t, s.Shutdown(defaultCtx))
	err := s.Run()
	assert.ErrorIs(t, err, server.ErrClosed, "should not run if already closed")
}

func TestServerHandlerCannotSendIfClosed(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, run := Server(t)
	s.Handle(defaultMsg().Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		// We wait for the closed state to send,
		// as we expect an error.
		time.Sleep(50 * time.Millisecond)
		if _, err := s.Send(msg); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()
	c, connect := Client(t, client.WithServerAddr(s.Addr()))
	connect()
	defer c.Close(defaultCtx)
	// We send the message, in order to invoke the handler.
	_, err := c.Send(defaultMsg())
	require.NoError(t, err)
	// While the handler is sleeping before sending,
	// we close the server, so changin its internal
	// state.
	require.NoError(t, s.Shutdown(defaultCtx))
	// The handler should try to send a message,
	// getting the expected error.
	arbiter.RequireErrorIs(server.ErrNotRunning)
}
