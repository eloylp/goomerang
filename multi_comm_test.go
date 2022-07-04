package goomerang_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestServerCanBroadCastMessages(t *testing.T) {
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
	brResult, err := s.BroadCast(defaultCtx, defaultMsg)
	require.NoError(t, err)
	require.Len(t, brResult, 2)
	require.Equal(t, 12, brResult[0].Size)
	require.NotEmpty(t, brResult[0].Duration)
	require.Equal(t, 12, brResult[1].Size)
	require.NotEmpty(t, brResult[1].Duration)
	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}

func TestClientsCanInterceptClosedConnection(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c1, connect1 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	defer c1.Close(defaultCtx)
	connect1()
	c2, connect2 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	defer c2.Close(defaultCtx)
	connect2()
	s.Shutdown(defaultCtx)

	_, err := c1.Send(defaultMsg)
	require.ErrorIs(t, err, client.ErrNotRunning, "expected client to intercept server close")
	_, err = c2.Send(defaultMsg)
	require.ErrorIs(t, err, client.ErrNotRunning, "expected client to intercept server close")
}

func TestServerSupportMultipleClients(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnErrorHook(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pingMsg, ok := msg.Payload.(*protos.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_FROM_CLIENT_" + pingMsg.Message)
		payload := &protos.MessageV1{Message: pingMsg.Message}
		_, err := ops.Send(message.New().SetPayload(payload))
		if err != nil {
			arbiter.ErrorHappened(err)
			return
		}
	}))
	run()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c1.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pongMsg, ok := msg.Payload.(*protos.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect1()
	defer c1.Close(defaultCtx)
	c2, connect2 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c2.Handle(defaultMsg.Payload, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pongMsg, ok := msg.Payload.(*protos.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect2()
	defer c2.Close(defaultCtx)

	_, err := c1.Send(message.New().SetPayload(&protos.MessageV1{Message: "1"}))
	require.NoError(t, err)
	_, err = c2.Send(message.New().SetPayload(&protos.MessageV1{Message: "2"}))
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_1")
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_2")
	arbiter.RequireHappened("CLIENT1_RECEIVED_FROM_SERVER_1")
	arbiter.RequireHappened("CLIENT2_RECEIVED_FROM_SERVER_2")
}
