//go:build unit

package messaging_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
)

func TestFQDN(t *testing.T) {
	m := &protos.MessageV1{}
	assert.Equal(t, "goomerang.example.MessageV1", messaging.FQDN(m))
}

func TestPackTimestamp(t *testing.T) {
	payload := &protos.MessageV1{}
	msg := &message.Message{
		Payload: payload,
	}
	_, pack, err := messaging.Pack(msg)
	require.NoError(t, err)
	unpack, err := messaging.UnPack(pack)
	require.NoError(t, err)
	now := time.Now().UnixMicro()
	packTime := unpack.Creation.AsTime().UnixMicro()
	assert.InDelta(t, now, packTime, 20000)
	assert.Less(t, packTime, now)
}

func TestFromFrame(t *testing.T) {
	inputMsg := &protos.MessageV1{Message: "This is a test message."}
	msgFQDN := string(inputMsg.ProtoReflect().Descriptor().FullName())
	inputMsgData, err := proto.Marshal(inputMsg)
	require.NoError(t, err)

	now := timestamppb.Now()
	header := message.Header{}
	header.Set("my-key", "my-value")
	size := int64(len(inputMsgData))

	frame := &protocol.Frame{
		Uuid:        "09AF",
		Kind:        msgFQDN,
		PayloadSize: size,
		IsSync:      true,
		Creation:    now,
		Payload:     inputMsgData,
		Headers:     header,
	}

	msgRegistry := message.Registry{}
	msgRegistry.Register(msgFQDN, inputMsg)

	msg, err := messaging.FromFrame(frame, msgRegistry)
	require.NoError(t, err)

	assert.Equal(t, "09AF", msg.Metadata.UUID)
	assert.Equal(t, msgFQDN, msg.Metadata.Kind)
	assert.Equal(t, size, int64(msg.Metadata.PayloadSize))
	assert.Equal(t, now.AsTime(), msg.Metadata.Creation)
	assert.Equal(t, true, msg.Metadata.IsSync)
	assert.Equal(t, "my-value", msg.Header.Get("my-key"))

	assert.Equal(t, inputMsg.Message, msg.Payload.(*protos.MessageV1).Message)
}
