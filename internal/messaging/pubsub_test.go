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

func TestMessageForPublish(t *testing.T) {
	msg := message.New().SetPayload(&protos.MessageV1{
		Message: "message for topic.a",
	})
	msg.SetHeader("k1", "v1")
	msg.SetHeader("k2", "v2")

	pubMsg, err := messaging.MessageForPublish("topic.a", msg)

	require.NoError(t, err)

	assert.Equal(t, "goomerang.protocol.PublishCommand", pubMsg.Metadata.Kind)
	assert.Equal(t, "v1", pubMsg.GetHeader("message-k1"))
	assert.Equal(t, "v2", pubMsg.GetHeader("message-k2"))
	payloadMsg := pubMsg.Payload.(*protocol.PublishCommand)
	assert.Equal(t, "topic.a", payloadMsg.Topic)
	assert.Len(t, payloadMsg.Message, 21)
	assert.Equal(t, "goomerang.example.MessageV1", payloadMsg.Kind)
}

func TestMessageFromPublish(t *testing.T) {
	pubCmd, clientProto := publishCommandFixture(t)

	mr := message.Registry{}
	mr.Register(messaging.FQDN(clientProto), clientProto)

	msgForPublish, err := messaging.MessageFromPublish(mr, pubCmd)

	require.NoError(t, err)
	assert.Equal(t, "v1", msgForPublish.GetHeader("k1"))
	assert.Equal(t, "v2", msgForPublish.GetHeader("k2"))
	assert.Equal(t, "goomerang.example.MessageV1", msgForPublish.Metadata.Kind)
	assert.Equal(t, clientProto.Message, msgForPublish.Payload.(*protos.MessageV1).Message)
}

func publishCommandFixture(t *testing.T) (*message.Message, *protos.MessageV1) {
	m := &protos.MessageV1{
		Message: "message for topic.a",
	}
	data, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	pubMsg := message.New().SetPayload(&protocol.PublishCommand{
		Topic:   "topic.a",
		Kind:    "goomerang.example.MessageV1",
		Message: data,
	})
	pubMsg.SetHeader("message-k1", "v1")
	pubMsg.SetHeader("message-k2", "v2")
	return pubMsg, m
}
