package goomerang_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestPingPongServer(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	ctx := context.Background()
	defer s.Shutdown(ctx)
	s.RegisterHandler(&testMessages.PingPong{}, func(s server.Ops, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		if err := s.Send(ctx, &testMessages.PingPong{
			Message: "pong",
		}); err != nil {
			return err
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
	err := c.Send(ctx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.AssertHappened("SERVER_RECEIVED_PING")
	arbiter.AssertHappened("CLIENT_RECEIVED_PONG")
}

func TestMultipleHandlersArePossibleInServer(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(context.Background())
	arbiter := NewArbiter(t)
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(ops server.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(ops server.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	h3 := func(ops server.Ops, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER3_CALLED")
		return nil
	}
	s.RegisterHandler(m, h, h2)
	s.RegisterHandler(m, h3)

	c := PrepareClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.Send(ctx, m)
	require.NoError(t, err)
	arbiter.AssertHappenedInOrder("HANDLER1_CALLED", "HANDLER2_CALLED")
	arbiter.AssertHappenedInOrder("HANDLER2_CALLED", "HANDLER3_CALLED")
}

func TestMultipleHandlersArePossibleInClient(t *testing.T) {
	ctx := context.Background()
	s := PrepareServer(t)
	defer s.Shutdown(ctx)
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

	err := s.Send(ctx, m)
	require.NoError(t, err)

	arbiter.AssertHappenedInOrder("HANDLER1_CALLED", "HANDLER2_CALLED")
	arbiter.AssertHappenedInOrder("HANDLER2_CALLED", "HANDLER3_CALLED")
}

func TestServerErrorHandler(t *testing.T) {
	arbiter := NewArbiter(t)
	ctx := context.Background()
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		if err != nil && err.Error() == "server handler err: a handler error" {
			arbiter.ItsAFactThat("ERROR_HANDLER_WORKS")
		}
	}))
	defer s.Shutdown(ctx)
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Ops, msg proto.Message) error {
		return errors.New("a handler error")
	})

	c1 := PrepareClient(t)
	defer c1.Close()

	err := c1.Send(ctx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.AssertHappened("ERROR_HANDLER_WORKS")
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
	arbiter.AssertHappened("SERVER_PROPERLY_CLOSED")
	arbiter.AssertHappened("CLIENT_PROPERLY_CLOSED")
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
	arbiter.AssertHappened("SERVER_PROPERLY_CLOSED")
	arbiter.AssertHappened("CLIENT_PROPERLY_CLOSED")
}

func TestServerSupportMultipleClients(t *testing.T) {
	ctx := context.Background()
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		arbiter.ErrorHappened(err)
	}))
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Ops, msg proto.Message) error {
		pingMsg, ok := msg.(*testMessages.PingPong)
		if !ok {
			return errors.New("cannot type assert message")
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_FROM_CLIENT_" + pingMsg.Message)
		return ops.Send(ctx, &testMessages.PingPong{Message: pingMsg.Message})
	})
	defer s.Shutdown(ctx)

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

	err := c1.Send(ctx, &testMessages.PingPong{Message: "1"})
	require.NoError(t, err)
	err = c2.Send(ctx, &testMessages.PingPong{Message: "2"})
	require.NoError(t, err)

	arbiter.AssertNoErrors()
	arbiter.AssertHappened("SERVER_RECEIVED_FROM_CLIENT_1")
	arbiter.AssertHappened("SERVER_RECEIVED_FROM_CLIENT_2")
	arbiter.AssertHappened("CLIENT1_RECEIVED_FROM_SERVER_1")
	arbiter.AssertHappened("CLIENT2_RECEIVED_FROM_SERVER_2")

}
