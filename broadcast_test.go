//go:build integration

package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestServerSideBroadcast(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, waitAndRun := PrepareServer(t)
	waitAndRun()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	defer c1.Close(defaultCtx)
	c1.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
	}))
	connect1()
	c2, connect2 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	defer c2.Close(defaultCtx)
	c2.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_SERVER_GREET")
	}))
	connect2()
	brResult, err := s.Broadcast(defaultCtx, defaultMsg)
	require.NoError(t, err)
	require.Len(t, brResult, 2)
	require.Equal(t, 12, brResult[0].Size)
	require.NotEmpty(t, brResult[0].Duration)
	require.Equal(t, 12, brResult[1].Size)
	require.NotEmpty(t, brResult[1].Duration)
	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}

func TestClientSideBroadCast(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, waitAndRun := PrepareServer(t, server.WithOnErrorHook(noErrorHook(arbiter)))
	s.RegisterMessage(defaultMsg.Payload)
	waitAndRun()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t, client.WithServerAddr(s.Addr()), client.WithOnErrorHook(noErrorHook(arbiter)))
	defer c1.Close(defaultCtx)
	c1.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
	}))
	connect1()
	c2, connect2 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	defer c2.Close(defaultCtx)
	c2.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_SERVER_GREET")
	}))
	connect2()

	brResult, err := c1.Broadcast(defaultMsg)
	require.NoError(t, err)
	assert.NotEmpty(t, brResult)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}
