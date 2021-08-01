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

func TestShutdownProcedure(t *testing.T) {
	s := PrepareServer(t)
	c := PrepareClient(t)

	require.NotPanics(t, func() {
		err := c.Close()
		require.NoError(t, err)
	})
	require.NotPanics(t, func() {
		err := s.Shutdown(context.Background())
		require.NoError(t, err)
	})
}
