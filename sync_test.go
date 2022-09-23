//go:build integration

package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
)

func TestSendSync(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := Server(t)
	defer s.Shutdown(defaultCtx)

	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		m := message.New().SetPayload(&protos.ReplyV1{Message: "pong !"})
		if _, err := s.Send(m); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()
	c, connect := Client(t, client.WithServerAddr(s.Addr()))
	connect()
	defer c.Close(defaultCtx)

	c.RegisterMessage(&protos.ReplyV1{})

	msg := message.New().SetPayload(&protos.MessageV1{Message: "ping"})

	payloadSize, reply, err := c.SendSync(defaultCtx, msg)
	require.NoError(t, err)
	require.NotEmpty(t, payloadSize)
	require.Equal(t, "pong !", reply.Payload.(*protos.ReplyV1).Message)
}
