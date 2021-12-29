package goomerang_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureClientSideInit(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	c := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	err := c.Close(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(context.Background())
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
}

func TestShutdownProcedureServerSideInit(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	c := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	defer c.Close(defaultCtx)
	err := s.Shutdown(context.Background())
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
}

func TestClientNormalClose(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	c := PrepareClient(t)

	require.NoError(t, c.Close(defaultCtx))
}

func TestClientCloseWhenServerClosed(t *testing.T) {
	s := PrepareServer(t)
	c := PrepareClient(t)

	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)

	require.ErrorIs(t, c.Close(defaultCtx), client.ErrServerDisconnected)
}

func TestServerSendWhenClientClosed(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	c := PrepareClient(t)

	require.NoError(t, c.Close(defaultCtx))

	msg := &testMessages.GreetV1{Message: "Hi!"}
	require.ErrorIs(t, s.Send(defaultCtx, msg), server.ErrClientDisconnected)
}
