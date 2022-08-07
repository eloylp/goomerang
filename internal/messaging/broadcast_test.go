package messaging_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
)

func TestMessageForBroadcast(t *testing.T) {
	msg := message.New().SetPayload(&protos.MessageV1{
		Message: "message for all !!",
	})
	msg.SetHeader("k1", "v1")
	msg.SetHeader("k2", "v2")

	brMsg, err := messaging.MessageForBroadcast(msg)

	require.NoError(t, err)

	assert.Equal(t, "goomerang.protocol.BroadcastCmd", brMsg.Metadata.Kind)
	assert.Equal(t, "v1", brMsg.GetHeader("message-k1"))
	assert.Equal(t, "v2", brMsg.GetHeader("message-k2"))
	payloadMsg := brMsg.Payload.(*protocol.BroadcastCmd)
	assert.Len(t, payloadMsg.Message, 20)
	assert.Equal(t, "goomerang.example.MessageV1", payloadMsg.Kind)
}

func TestMessageFromBroadcast(t *testing.T) {
	brCmd, clientProto := broadcastCmdFixture(t)

	mr := message.Registry{}
	mr.Register(messaging.FQDN(clientProto), clientProto)

	msgForPublish, err := messaging.MessageFromBroadcast(mr, brCmd)

	require.NoError(t, err)
	assert.Equal(t, "v1", msgForPublish.GetHeader("k1"))
	assert.Equal(t, "v2", msgForPublish.GetHeader("k2"))
	assert.Equal(t, "goomerang.example.MessageV1", msgForPublish.Metadata.Kind)
	assert.Equal(t, clientProto.Message, msgForPublish.Payload.(*protos.MessageV1).Message)
}

func broadcastCmdFixture(t *testing.T) (*message.Message, *protos.MessageV1) {
	m := &protos.MessageV1{
		Message: "message for all !!",
	}
	data, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	brMsg := message.New().SetPayload(&protocol.BroadcastCmd{
		Kind:    "goomerang.example.MessageV1",
		Message: data,
	})
	brMsg.SetHeader("message-k1", "v1")
	brMsg.SetHeader("message-k2", "v2")
	return brMsg, m
}
