package goomerang_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestPingPongServer(t *testing.T) {
	const (
		serverReceivedPing = "SERVER_RECEIVED_PING"
		clientReceivedPong = "CLIENT_RECEIVED_PONG"
	)
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	ctx := context.Background()
	defer s.Shutdown(ctx)
	s.RegisterHandler(&testMessages.PingPong{}, func(s server.Opts, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		if err := s.Send(ctx, &testMessages.PingPong{
			Message: "pong",
		}); err != nil {
			return err
		}
		arbiter.ItsAFactThat(serverReceivedPing)
		return nil
	})

	c := PrepareClient(t)
	defer c.Close()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c.RegisterHandler(&testMessages.PingPong{}, func(c client.Ops, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		arbiter.ItsAFactThat(clientReceivedPong)
		wg.Done()
		return nil
	})

	err := c.Send(ctx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)

	wg.Wait()

	arbiter.AssertHappened(serverReceivedPing)
	arbiter.AssertHappened(clientReceivedPong)
}

func TestMultipleHandlersArePossible(t *testing.T) {
	s := PrepareServer(t)
	arbiter := NewArbiter(t)
	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(serverOpts server.Opts, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(serverOpts server.Opts, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	h3 := func(serverOpts server.Opts, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER3_CALLED")
		wg.Done()
		return nil
	}
	s.RegisterHandler(m, h, h2)
	s.RegisterHandler(m, h3)

	c := PrepareClient(t)
	defer c.Close()

	ctx := context.Background()
	err := c.Send(ctx, m)
	require.NoError(t, err)

	wg.Wait()

	arbiter.AssertHappenedInOrder("HANDLER1_CALLED", "HANDLER2_CALLED")
	arbiter.AssertHappenedInOrder("HANDLER2_CALLED", "HANDLER3_CALLED")
}
