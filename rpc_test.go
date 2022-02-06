package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/internal/test"
)

func TestRPC(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	s.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(ops message.Sender, msg *message.Request) {
		if err := ops.Send(defaultCtx, &testMessages.PingPong{Message: "pong !"}); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterMessage(&testMessages.PingPong{})

	msg := &testMessages.PingPong{Message: "ping"}

	reply, err := c.RPC(defaultCtx, msg)
	require.NoError(t, err)
	require.Equal(t, "pong !", reply.Payload.(*testMessages.PingPong).Message)
}
