//go:build integration

package goomerang_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
)

func TestUserCanConfigureBroadcastHandler(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	// The server should have the broadcast message previously registered.
	s.Handle(&protos.MessageV1{}, nilHandler)
	// Register the broadcast message envelope type at server
	s.Handle(&protos.BroadCastV1{}, message.HandlerFunc(func(sender message.Sender, iMsg *message.Message) {
		broadMsg := iMsg.Payload.(*protos.BroadCastV1)

		oMsg, err := s.Registry().Message(broadMsg.Kind)
		if err != nil {
			arbiter.ErrorHappened(err)
			return
		}
		if err := proto.Unmarshal(broadMsg.BroadcastedMessage, oMsg); err != nil {
			arbiter.ErrorHappened(err)
			return
		}
		_, err = s.BroadCast(context.TODO(), message.New().SetPayload(oMsg))
		if err != nil {
			arbiter.ErrorHappened(err)
			return
		}
	}))
	run()
	defer s.Shutdown(defaultCtx)

	// Prepare the client to receive the broadcast message.
	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_BROADCAST_MESSAGE")
	}))
	connect()
	defer c.Close(defaultCtx)

	// Request to server the message broadcast
	m := &protos.MessageV1{
		Message: "a message which is going to be broadcast",
	}
	data, err := proto.Marshal(m)
	require.NoError(t, err)
	_, err = c.Send(message.New().SetPayload(&protos.BroadCastV1{
		BroadcastedMessage: data,
		Kind:               string(m.ProtoReflect().Descriptor().FullName()),
	}))
	require.NoError(t, err)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT_RECEIVED_BROADCAST_MESSAGE")
}
