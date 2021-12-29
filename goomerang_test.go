package goomerang_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"

	"go.eloylp.dev/goomerang/client"
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
