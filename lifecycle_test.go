package goomerang_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

func TestShutdownProcedureClientSideInit(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	run()
	c, connect := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	connect()
	err := c.Close(defaultCtx)
	require.NoError(t, err)
	err = s.Shutdown(context.Background())
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
}

func TestShutdownProcedureServerSideInit(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	run()
	c, connect := PrepareClient(t, client.WithOnCloseHook(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	connect()
	defer c.Close(defaultCtx)

	require.NoError(t, s.Shutdown(defaultCtx))
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
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

	require.ErrorIs(t, c.Close(defaultCtx), client.ErrServerDisconnected)
}
