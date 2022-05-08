package goomerang_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestServerCanBroadCastMessages(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, waitAndRun := PrepareServer(t)
	waitAndRun()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	c1.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
	}))
	connect1()
	c2, connect2 := PrepareClient(t)
	defer c2.Close(defaultCtx)
	c2.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_SERVER_GREET")
	}))
	connect2()
	payloadSize, sentCount, err := s.BroadCast(defaultCtx, defaultMsg)
	require.NoError(t, err)
	require.NotEmpty(t, payloadSize)
	require.Equal(t, 2, sentCount)
	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}

func TestClientsCanInterceptClosedConnection(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c1, connect1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	connect1()
	c2, connect2 := PrepareClient(t)
	defer c2.Close(defaultCtx)
	connect2()
	s.Shutdown(defaultCtx)

	_, err := c1.Send(defaultMsg)
	require.EqualErrorf(t, err, "client: not running", "expected client to intercept server close")
	_, err = c2.Send(defaultMsg)
	require.EqualErrorf(t, err, "client: not running", "expected client to intercept server close")
}

func TestServerSupportMultipleClients(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnErrorHook(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pingMsg, ok := msg.Payload.(*testMessages.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_FROM_CLIENT_" + pingMsg.Message)
		payload := &testMessages.MessageV1{Message: pingMsg.Message}
		_, err := ops.Send(message.New().SetPayload(payload))
		if err != nil {
			arbiter.ErrorHappened(err)
			return
		}
	}))
	run()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t)
	c1.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pongMsg, ok := msg.Payload.(*testMessages.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect1()
	defer c1.Close(defaultCtx)
	c2, connect2 := PrepareClient(t)
	c2.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(ops message.Sender, msg *message.Message) {
		pongMsg, ok := msg.Payload.(*testMessages.MessageV1)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect2()
	defer c2.Close(defaultCtx)

	_, err := c1.Send(message.New().SetPayload(&testMessages.MessageV1{Message: "1"}))
	require.NoError(t, err)
	_, err = c2.Send(message.New().SetPayload(&testMessages.MessageV1{Message: "2"}))
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_1")
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_2")
	arbiter.RequireHappened("CLIENT1_RECEIVED_FROM_SERVER_1")
	arbiter.RequireHappened("CLIENT2_RECEIVED_FROM_SERVER_2")
}
