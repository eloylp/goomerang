package goomerang_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureClientSideInit(t *testing.T) {
	serverArbiter := test.NewArbiter(t)
	clientArbiter := test.NewArbiter(t)

	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		serverArbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}),
		server.WithOnStatusChangeHook(statusChangesHook(serverArbiter, "server")),
		server.WithOnErrorHook(noErrorHook(serverArbiter)),
	)
	run()

	c, connect := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnCloseHook(func() {
			clientArbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
		}),
		client.WithOnStatusChangeHook(statusChangesHook(clientArbiter, "client")),
		client.WithOnErrorHook(noErrorHook(clientArbiter)),
	)
	connect()

	err := c.Close(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(context.Background())
	require.NoError(t, err)

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

func TestClientNormalClose(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()
	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	connect()
	require.NoError(t, c.Close(defaultCtx))
}

func TestClientShouldBeClosedWhenServerCloses(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
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

	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	connect()

	require.NoError(t, c.Close(defaultCtx))
	require.NoError(t, c.Connect(defaultCtx))
	require.NoError(t, c.Close(defaultCtx))
}

func TestClientCannotConnectForSecondTime(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	connect()
	defer c.Close(defaultCtx)

	assert.ErrorIs(t, c.Connect(defaultCtx), client.ErrAlreadyRunning)
}

func TestClientCannotSendMessagesIfNotRunning(t *testing.T) {
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	run()

	// Cannot send if not connected
	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
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

	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	assert.ErrorIs(t, c.Close(defaultCtx), client.ErrNotRunning, "should not shutdown if not running")
	connect()
	require.NoError(t, c.Close(defaultCtx))

	assert.ErrorIs(t, c.Close(defaultCtx), client.ErrNotRunning, "should not shutdown if already closed")
}

func TestClientHandlerCannotSendIfClosed(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	run()
	defer s.Shutdown(defaultCtx)
	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		// Let's wait the close before sending.
		// We expect an error here.
		time.Sleep(50 * time.Millisecond)
		if _, err := s.Send(msg); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	connect()

	// We provoke the handler execution of the client
	// with a broadcast message.
	_, err := s.BroadCast(defaultCtx, defaultMsg)
	require.NoError(t, err)

	// We close the client, now the internal status changed.
	require.NoError(t, c.Close(defaultCtx))
	// At this point the wait time of the handler has expired,
	// so we check it tried to send a message back in a closed state,
	// causing the error.
	arbiter.RequireErrorIs(client.ErrNotRunning)
}
