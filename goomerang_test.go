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

var defaultCtx = context.Background()

func TestPingPongServer(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	s.RegisterHandler(&testMessages.PingPong{}, func(s server.Ops, msg proto.Message) *server.HandlerError {
		_ = msg.(*testMessages.PingPong)
		if err := s.Send(defaultCtx, &testMessages.PingPong{
			Message: "pong",
		}); err != nil {
			return server.NewHandlerError("sd")
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_PING")
		return nil
	})

	c := PrepareClient(t)
	defer c.Close()

	c.RegisterHandler(&testMessages.PingPong{}, func(c client.Ops, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_PONG")
		return nil
	})
	err := c.Send(defaultCtx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_RECEIVED_PING")
	arbiter.RequireHappened("CLIENT_RECEIVED_PONG")
}

func TestMultipleHandlersArePossibleInServer(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(context.Background())
	arbiter := NewArbiter(t)
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(ops server.Ops, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(ops server.Ops, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	h3 := func(ops server.Ops, msg proto.Message) *server.HandlerError {
		arbiter.ItsAFactThat("HANDLER3_CALLED")
		return nil
	}
	s.RegisterHandler(m, h, h2)
	s.RegisterHandler(m, h3)

	c := PrepareClient(t)
	defer c.Close()

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
	defer c.Close()
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(opts client.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(ops client.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	h3 := func(ops client.Ops, msg proto.Message) error {
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

func TestServerErrorHandler(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		if err != nil {
			arbiter.ItsAFactThat("ERROR_HANDLER_WORKS")
		}
	}))
	defer s.Shutdown(defaultCtx)
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Ops, msg proto.Message) *server.HandlerError {
		return server.NewHandlerError("a handler error")
	})

	c1 := PrepareClient(t)
	defer c1.Close()

	err := c1.Send(defaultCtx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.RequireHappened("ERROR_HANDLER_WORKS")
}

func TestShutdownProcedureClientSideInit(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithOnCloseHandler(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	c := PrepareClient(t, client.WithOnCloseHandler(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	err := c.Close()
	require.NoError(t, err)
	err = s.Shutdown(context.Background())
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
}

func TestShutdownProcedureServerSideInit(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithOnCloseHandler(func() {
		arbiter.ItsAFactThat("SERVER_PROPERLY_CLOSED")
	}))
	c := PrepareClient(t, client.WithOnCloseHandler(func() {
		arbiter.ItsAFactThat("CLIENT_PROPERLY_CLOSED")
	}))
	defer c.Close()
	err := s.Shutdown(context.Background())
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_PROPERLY_CLOSED")
	arbiter.RequireHappened("CLIENT_PROPERLY_CLOSED")
}

func TestServerSupportMultipleClients(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Ops, msg proto.Message) *server.HandlerError {
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
	c1.RegisterHandler(&testMessages.PingPong{}, func(ops client.Ops, msg proto.Message) error {
		pongMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return errors.New("cannot type assert message")
		}
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_FROM_SERVER_" + pongMsg.Message)
		return nil
	})
	defer c1.Close()
	c2 := PrepareClient(t)
	c2.RegisterHandler(&testMessages.PingPong{}, func(ops client.Ops, msg proto.Message) error {
		pongMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return errors.New("cannot type assert message")
		}
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_FROM_SERVER_" + pongMsg.Message)
		return nil
	})
	defer c2.Close()

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

func TestServerCanBroadCastMessages(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	c1 := PrepareClient(t)
	defer c1.Close()
	c1.RegisterHandler(&testMessages.GreetV1{}, func(ops client.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_SERVER_GREET")
		return nil
	})

	c2 := PrepareClient(t)
	defer c2.Close()
	c2.RegisterHandler(&testMessages.GreetV1{}, func(ops client.Ops, msg proto.Message) error {
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
	defer c1.Close()
	c2 := PrepareClient(t)
	defer c2.Close()

	s.Shutdown(defaultCtx)

	msg := &testMessages.GreetV1{Message: "Hi!"}

	require.ErrorIs(t, c1.Send(defaultCtx, msg), client.ErrServerDisconnected)
	require.ErrorIs(t, c2.Send(defaultCtx, msg), client.ErrServerDisconnected)
}

func TestClientNormalClose(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	c := PrepareClient(t)

	require.NoError(t, c.Close())
}

func TestClientCloseWhenServerClosed(t *testing.T) {
	s := PrepareServer(t)
	c := PrepareClient(t)

	err := s.Shutdown(defaultCtx)
	require.NoError(t, err)

	require.ErrorIs(t, c.Close(), client.ErrServerDisconnected)
}

func TestServerSendWhenClientClosed(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	c := PrepareClient(t)

	require.NoError(t, c.Close())

	msg := &testMessages.GreetV1{Message: "Hi!"}
	require.ErrorIs(t, s.Send(defaultCtx, msg), server.ErrClientDisconnected)
}

func TestRequestReplyPattern(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	h1 := func(ops server.Ops, msg proto.Message) *server.HandlerError {
		if err := ops.Send(defaultCtx, &testMessages.PingPong{Message: "pong !"}); err != nil {
			return server.NewHandlerError(err.Error())
		}
		return nil
	}
	h2 := func(ops server.Ops, msg proto.Message) *server.HandlerError {
		return server.NewHandlerErrorWith("happened at handler !", 500)
	}

	s.RegisterHandler(&testMessages.PingPong{}, h1, h2)

	c := PrepareClient(t)
	defer c.Close()

	c.RegisterMessage(&testMessages.PingPong{})

	msg := &testMessages.PingPong{Message: "ping"}

	replies, err := c.RPC(defaultCtx, msg)
	require.NoError(t, err)

	repMsg1 := replies.First()
	require.Nil(t, repMsg1.Err)
	require.Equal(t, "pong !", repMsg1.Message.(*testMessages.PingPong).Message)

	repMsg2 := replies.Index(1)
	require.Equal(t, "happened at handler !", repMsg2.Err.Message)
	require.Equal(t, int32(500), repMsg2.Err.Code)
}
