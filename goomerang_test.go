package goomerang_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/server"
)

var defaultCtx = context.Background()

func TestPingPongServer(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	s.RegisterHandler(&testMessages.PingPong{}, func(s server.Sender, msg proto.Message) *server.HandlerError {
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
	defer c.Close(defaultCtx)

	c.RegisterHandler(&testMessages.PingPong{}, func(c client.Sender, msg proto.Message) error {
		_ = msg.(*testMessages.PingPong)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_PONG")
		return nil
	})
	err := c.Send(defaultCtx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.RequireHappened("SERVER_RECEIVED_PING")
	arbiter.RequireHappened("CLIENT_RECEIVED_PONG")
}

func TestServerErrorHandler(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t, server.WithErrorHandler(func(err error) {
		if err != nil {
			arbiter.ItsAFactThat("ERROR_HANDLER_WORKS")
		}
	}))
	defer s.Shutdown(defaultCtx)
	s.RegisterHandler(&testMessages.PingPong{}, func(ops server.Sender, msg proto.Message) *server.HandlerError {
		return server.NewHandlerError("a handler error")
	})

	c1 := PrepareClient(t)
	defer c1.Close(defaultCtx)

	err := c1.Send(defaultCtx, &testMessages.PingPong{Message: "ping"})
	require.NoError(t, err)
	arbiter.RequireHappened("ERROR_HANDLER_WORKS")
}

func TestClientMessageProcessedHandler(t *testing.T) {
	arbiter := NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	msg := &testMessages.PingPong{}
	c := PrepareClient(t, client.WithOnMessageProcessedHandler(func(name string, duration time.Duration) {
		if name == message.FQDN(msg) {
			arbiter.ItsAFactThat("MESSAGE_PROCESSED_HANDLER_RECEIVED_NAME")
		}
		if duration.Seconds() != 0 {
			arbiter.ItsAFactThat("MESSAGE_PROCESSED_HANDLER_RECEIVED_DURATION")
		}
	}))
	defer c.Close(defaultCtx)
	c.RegisterHandler(msg, func(ops client.Sender, msg proto.Message) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	err := s.Send(defaultCtx, msg)
	assert.NoError(t, err)

	arbiter.RequireHappened("MESSAGE_PROCESSED_HANDLER_RECEIVED_NAME")
	arbiter.RequireHappened("MESSAGE_PROCESSED_HANDLER_RECEIVED_DURATION")
}
