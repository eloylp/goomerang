package goomerang_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"sync"
	"testing"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestPingPongServer(t *testing.T) {
	// Just prepare our assertion testMessages
	arbiter := NewArbiter(t)

	wg := &sync.WaitGroup{}
	// As all the processing is async in other goroutines,
	// we will use this sync primitive in order to wait the
	// end of the processing.
	wg.Add(1)

	s := PrepareServer(t)
	ctx := context.Background()
	defer s.Shutdown(ctx)

	// Register handler at server
	s.RegisterHandler(&testMessages.PingPong{}, serverHandler(arbiter, ctx))

	// Client side
	c := PrepareClient(t)
	defer c.Close()

	// Register handler at client (as this is a bidirectional communication)
	c.RegisterHandler(&testMessages.PingPong{}, clientHandler(arbiter, wg))
	err := c.Send(ctx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)

	wg.Wait()

	arbiter.AssertHappened(serverReceivedPing)
	arbiter.AssertHappened(clientReceivedPong)
}

func clientHandler(arbiter *Arbiter, wg *sync.WaitGroup) client.Handler {
	return func(c client.Ops, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		arbiter.ItsAFactThat(clientReceivedPong)
		wg.Done()
		return nil
	}
}

func serverHandler(arbiter *Arbiter, ctx context.Context) server.Handler {
	return func(s server.Opts, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		if err := s.Send(ctx, &testMessages.PingPong{ // As all the processing is async in other goroutines, we will use this sync primitive in order to wait the end of the processing.
			Message: "pong",
		}); err != nil {
			return err
		}
		arbiter.ItsAFactThat(serverReceivedPing)
		return nil
	}
}

func TestMultipleHandlersArePossible(t *testing.T) {
	s := PrepareServer(t)
	arbiter := NewArbiter(t)

	wg := sync.WaitGroup{}
	wg.Add(1)

	m := &testMessages.GreetV1{Message: "Hi !"}
	h := func(serverOpts server.Opts, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER1_CALLED")
		return nil
	}
	h2 := func(serverOpts server.Opts, msg proto.Message) error {
		arbiter.ItsAFactThat("HANDLER2_CALLED")
		return nil
	}
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
