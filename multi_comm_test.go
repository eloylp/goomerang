package goomerang_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestServerCanBroadCastMessages(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	c1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	c1.RegisterHandler(&testMessages.GreetV1{}, func(ops client.Sender, msg proto.Message) error {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
		return nil
	})

	c2 := PrepareClient(t)
	defer c2.Close(defaultCtx)
	c2.RegisterHandler(&testMessages.GreetV1{}, func(ops client.Sender, msg proto.Message) error {
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_SERVER_GREET")
		return nil
	})

	err := s.Send(defaultCtx, &testMessages.GreetV1{Message: "Hi!"})
	require.NoError(t, err)

	arbiter.RequireHappened("CLIENT1_RECEIVED_SERVER_GREET")
	arbiter.RequireHappened("CLIENT2_RECEIVED_SERVER_GREET")
}

func TestServerShutdownIsPropagatedToAllClients(t *testing.T) {
	s := PrepareServer(t)

	c1 := PrepareClient(t)
	defer c1.Close(defaultCtx)
	c2 := PrepareClient(t)
	defer c2.Close(defaultCtx)

	s.Shutdown(defaultCtx)

	msg := &testMessages.GreetV1{Message: "Hi!"}

	require.ErrorIs(t, c1.Send(defaultCtx, msg), client.ErrServerDisconnected)
	require.ErrorIs(t, c2.Send(defaultCtx, msg), client.ErrServerDisconnected)
}

func TestServerSupportMultipleClients(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Sender, msg proto.Message) *server.HandlerError {
		pingMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return server.NewHandlerError("cannot type assert message")
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_FROM_CLIENT_" + pingMsg.Message)
		err := ops.Send(defaultCtx, &testMessages.PingPong{Message: pingMsg.Message})
		if err != nil {
			return server.NewHandlerError(err.Error())
		}
		return nil
	})
	defer s.Shutdown(defaultCtx)

	c1 := PrepareClient(t)
	c1.RegisterHandler(&testMessages.PingPong{}, func(ops client.Sender, msg proto.Message) error {
		pongMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return errors.New("cannot type assert message")
		}
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_FROM_SERVER_" + pongMsg.Message)
		return nil
	})
	defer c1.Close(defaultCtx)
	c2 := PrepareClient(t)
	c2.RegisterHandler(&testMessages.PingPong{}, func(ops client.Sender, msg proto.Message) error {
		pongMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return errors.New("cannot type assert message")
		}
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_FROM_SERVER_" + pongMsg.Message)
		return nil
	})
	defer c2.Close(defaultCtx)

	err := c1.Send(defaultCtx, &testMessages.PingPong{Message: "1"})
	require.NoError(t, err)
	err = c2.Send(defaultCtx, &testMessages.PingPong{Message: "2"})
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_1")
	arbiter.RequireHappened("SERVER_RECEIVED_FROM_CLIENT_2")
	arbiter.RequireHappened("CLIENT1_RECEIVED_FROM_SERVER_1")
	arbiter.RequireHappened("CLIENT2_RECEIVED_FROM_SERVER_2")
}

func TestMultipleHandlersArePossibleInServer(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(context.Background())
	arbiter := NewArbiter(t)
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(ops server.Sender, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(ops server.Sender, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	h3 := func(ops server.Sender, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER3_CALLED")
		return nil
	}
	s.RegisterHandler(m, h, h2)
	s.RegisterHandler(m, h3)

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	err := c.Send(defaultCtx, m)
	require.NoError(t, err)
	arbiter.RequireHappenedInOrder("HANDLER1_CALLED", "HANDLER2_CALLED")
	arbiter.RequireHappenedInOrder("HANDLER2_CALLED", "HANDLER3_CALLED")
}

func TestMultipleHandlersArePossibleInClient(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	arbiter := NewArbiter(t)
	c := PrepareClient(t)
	defer c.Close(defaultCtx)
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(opts client.Sender, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(ops client.Sender, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	h3 := func(ops client.Sender, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER3_CALLED")
		return nil
	}
	c.RegisterHandler(m, h, h2)
	c.RegisterHandler(m, h3)

	err := s.Send(defaultCtx, m)
	require.NoError(t, err)

	arbiter.RequireHappenedInOrder("HANDLER1_CALLED", "HANDLER2_CALLED")
	arbiter.RequireHappenedInOrder("HANDLER2_CALLED", "HANDLER3_CALLED")
}
