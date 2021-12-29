package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestRPC(t *testing.T) {
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	h1 := func(ops server.Sender, msg proto.Message) *server.HandlerError {
		if err := ops.Send(defaultCtx, &testMessages.PingPong{Message: "pong !"}); err != nil {
			return server.NewHandlerError(err.Error())
		}
		return nil
	}
	h2 := func(ops server.Sender, msg proto.Message) *server.HandlerError {
		return server.NewHandlerErrorWith("happened at handler !", 500)
	}

	s.RegisterHandler(&testMessages.PingPong{}, h1, h2)

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterMessage(&testMessages.PingPong{})

	msg := &testMessages.PingPong{Message: "ping"}

	replies, err := c.RPC(defaultCtx, msg)
	require.NoError(t, err)

	repMsg1 := replies.First()
	require.Nil(t, repMsg1.Err)
	require.Equal(t, "pong !", repMsg1.Message.(*testMessages.PingPong).Message)

	repMsg2 := replies.Index(1)
	require.Equal(t, "happened at handler !", repMsg2.Err.Message)
	require.Equal(t, int32(500), repMsg2.Err.Code)
}
