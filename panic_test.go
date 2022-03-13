package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	test "go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestClientHandlesPanic(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	run()
	defer s.Shutdown(defaultCtx)
	c, connect := PrepareClient(t, client.WithOnErrorHook(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	c.RegisterHandler(&testMessages.GreetV1{}, message.HandlerFunc(func(_ message.Sender, _ *message.Message) {
		panic("handler panic !")
	}))
	connect()
	defer c.Close(defaultCtx)

	require.NoError(t, s.Send(defaultCtx, &message.Message{
		Payload: &testMessages.GreetV1{Message: "Hi! you are going to panic :D"},
	}))
	arbiter.RequireError("goomerang: client: panic: handler panic !")
}

func TestServerHandlesPanic(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnErrorHook(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.GreetV1{}, message.HandlerFunc(func(_ message.Sender, _ *message.Message) {
		panic("handler panic !")
	}))
	run()
	defer s.Shutdown(defaultCtx)

	c, connect := PrepareClient(t)
	connect()
	defer c.Close(defaultCtx)

	require.NoError(t, c.Send(defaultCtx, &message.Message{
		Payload: &testMessages.GreetV1{Message: "Hi! you are going to panic :D"},
	}))
	arbiter.RequireError("goomerang: server: panic: handler panic !")
}
