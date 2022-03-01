package goomerang_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

func TestServerCanBroadCastMessages(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, waitAndRun := PrepareServer(t)
	waitAndRun()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	c1.RegisterHandler(&testMessages.GreetV1{}, message.HandlerFunc(func(ops message.Sender, msg *goomerang.Message) {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
	}))
	connect1()
	c2, connect2 := PrepareClient(t)
	defer c2.Close(defaultCtx)
	c2.RegisterHandler(&testMessages.GreetV1{}, message.HandlerFunc(func(ops message.Sender, msg *goomerang.Message) {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_SERVER_GREET")
	}))
	connect2()
	err := s.Send(defaultCtx, &goomerang.Message{Payload: &testMessages.GreetV1{Message: "Hi!"}})
	require.NoError(t, err)

	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}

func TestServerShutdownIsPropagatedToAllClients(t *testing.T) {
	s, run := PrepareServer(t)
	run()
	c1, connect1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	connect1()
	c2, connect2 := PrepareClient(t)
	defer c2.Close(defaultCtx)
	connect2()
	s.Shutdown(defaultCtx)

	msg := &goomerang.Message{Payload: &testMessages.GreetV1{Message: "Hi!"}}

	require.ErrorIs(t, c1.Send(defaultCtx, msg), client.ErrServerDisconnected)
	require.ErrorIs(t, c2.Send(defaultCtx, msg), client.ErrServerDisconnected)
}

func TestServerSupportMultipleClients(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnErrorHook(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(ops message.Sender, msg *goomerang.Message) {
		pingMsg, ok := msg.Payload.(*testMessages.PingPong)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_FROM_CLIENT_" + pingMsg.Message)
		err := ops.Send(defaultCtx, &goomerang.Message{Payload: &testMessages.PingPong{Message: pingMsg.Message}})
		if err != nil {
			arbiter.ErrorHappened(err)
			return
		}
	}))
	run()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t)
	c1.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(ops message.Sender, msg *goomerang.Message) {
		pongMsg, ok := msg.Payload.(*testMessages.PingPong)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect1()
	defer c1.Close(defaultCtx)
	c2, connect2 := PrepareClient(t)
	c2.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(ops message.Sender, msg *goomerang.Message) {
		pongMsg, ok := msg.Payload.(*testMessages.PingPong)
		if !ok {
			arbiter.ErrorHappened(errors.New("cannot type assert message"))
			return
		}
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_FROM_SERVER_" + pongMsg.Message)
	}))
	connect2()
	defer c2.Close(defaultCtx)

	err := c1.Send(defaultCtx, &goomerang.Message{Payload: &testMessages.PingPong{Message: "1"}})
	require.NoError(t, err)
	err = c2.Send(defaultCtx, &goomerang.Message{Payload: &testMessages.PingPong{Message: "2"}})
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_1")
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_2")
	arbiter.RequireHappened("CLIENT1_RECEIVED_FROM_SERVER_1")
	arbiter.RequireHappened("CLIENT2_RECEIVED_FROM_SERVER_2")
}
