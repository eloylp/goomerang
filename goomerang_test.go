package goomerang_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	messages2 "go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/kit/test"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang"
)

const (
	serverAddr = "0.0.0.0:3000"
)

func TestPingPongServer(t *testing.T) {

	// Just prepare our assertion message
	arbiter := NewArbiter()

	wg := &sync.WaitGroup{}
	// As all the processing is async in other goroutines,
	// we will use this sync primitive in order to wait the
	// end of the processing.
	wg.Add(1)

	// Create the test server
	s, err := goomerang.NewServer(goomerang.WithListenAddr(serverAddr))
	require.NoError(t, err)
	ctx := context.Background()

	go s.Run()
	test.WaitTCPService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	defer s.Shutdown(ctx)

	// Register handler at server
	s.RegisterHandler(&messages2.PingPong{}, serverHandler(arbiter, ctx))
	// Create client
	c, err := goomerang.NewClient(goomerang.WithTargetServer(serverAddr))
	require.NoError(t, err)
	err = c.Connect(ctx)
	require.NoError(t, err)
	defer c.Close()

	// Register handler at client (as this is a bidirectional communication)
	c.RegisterHandler(&messages2.PingPong{}, clientHandler(arbiter, wg))
	err = c.Send(ctx, &messages2.PingPong{Message: "ping"})
	require.NoError(t, err)

	wg.Wait()

	assert.True(t, arbiter.AssertHappened(serverReceivedPing))
	assert.True(t, arbiter.AssertHappened(clientReceivedPong))
}

func clientHandler(arbiter *Arbiter, wg *sync.WaitGroup) goomerang.ClientHandler {
	return func(c *goomerang.Client, msg proto.Message) error {
		_ = msg.(*messages2.PingPong)
		arbiter.ItsAFactThat(clientReceivedPong)
		wg.Done()
		return nil
	}
}

func serverHandler(arbiter *Arbiter, ctx context.Context) goomerang.ServerHandler {
	return func(s *goomerang.Server, msg proto.Message) error {
		_ = msg.(*messages2.PingPong)
		if err := s.Send(ctx, &messages2.PingPong{ // As all the processing is async in other goroutines, we will use this sync primitive in order to wait the end of the processing.
			Message: "pong",
		}); err != nil {
			return err
		}
		arbiter.ItsAFactThat(serverReceivedPing)
		return nil
	}
}
