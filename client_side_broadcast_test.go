package goomerang_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
)

func TestUserCanConfigureBroadcastHandler(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	// The server should have the broadcast message previously registered.
	s.Handle(&testMessages.MessageV1{}, nilHandler)
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
		_ = msg.Payload.(*testMessages.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_BROADCAST_MESSAGE")
	}))
	connect()
	defer c.Close(defaultCtx)

	// Request to server the message broadcast
	m := &testMessages.MessageV1{
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
