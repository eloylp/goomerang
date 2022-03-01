package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/protocol"
	"go.eloylp.dev/goomerang/internal/message/test"
)

func TestFQDN(t *testing.T) {
	m := &test.GreetV1{}
	assert.Equal(t, "goomerang.test.GreetV1", message.FQDN(m))
}

func TestPackTimestamp(t *testing.T) {
	payload := &test.GreetV1{}
	msg := &goomerang.Message{
		Payload: payload,
	}
	pack, err := message.Pack(msg)
	require.NoError(t, err)
	unpack, err := message.UnPack(pack)
	require.NoError(t, err)
	now := time.Now().UnixMicro()
	packTime := unpack.Creation.AsTime().UnixMicro()
	assert.InDelta(t, now, packTime, 1000)
	assert.Less(t, packTime, now)
}

func TestFromFrame(t *testing.T) {
	inputMsg := &test.PingPong{}
	msgFQDN := string(inputMsg.ProtoReflect().Descriptor().FullName())
	inputMsgData, err := proto.Marshal(inputMsg)
	require.NoError(t, err)

	now := timestamppb.Now()
	header := goomerang.Header{}
	header.Add("my-key", "my-value")

	frame := &protocol.Frame{
		Uuid:     "09AF",
		Type:     msgFQDN,
		IsRpc:    true,
		Creation: now,
		Payload:  inputMsgData,
		Headers:  header,
	}

	msgRegistry := message.Registry{}
	msgRegistry.Register(msgFQDN, inputMsg)

	msg, err := message.FromFrame(frame, msgRegistry)
	require.NoError(t, err)

	assert.Equal(t, "09AF", msg.Metadata.UUID)
	assert.Equal(t, msgFQDN, msg.Metadata.Type)
	assert.Equal(t, now.AsTime(), msg.Metadata.Creation)
	assert.Equal(t, true, msg.Metadata.IsRPC)
	assert.Equal(t, "my-value", msg.Header.Get("my-key"))

	assert.Equal(t, inputMsg, msg.Payload)
}
