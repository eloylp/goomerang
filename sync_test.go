package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
)

func TestSync(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	s.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		m := message.New().SetPayload(&testMessages.MessageV1{Message: "pong !"})
		if _, err := s.Send(m); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()
	c, connect := PrepareClient(t)
	connect()
	defer c.Close(defaultCtx)

	c.RegisterMessage(&testMessages.MessageV1{})

	msg := message.New().SetPayload(&testMessages.MessageV1{Message: "ping"})

	payloadSize, reply, err := c.SendSync(defaultCtx, msg)
	require.NoError(t, err)
	require.NotEmpty(t, payloadSize)
	require.Equal(t, "pong !", reply.Payload.(*testMessages.MessageV1).Message)
}
