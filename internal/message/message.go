package message

import (
	"go.eloylp.dev/goomerang/internal/message/protocol"
	"google.golang.org/protobuf/proto"
)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

func PackMessage(msg proto.Message) ([]byte, error) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	envelope := &protocol.Frame{
		Type:    FQDN(msg),
		Payload: payload,
	}
	bytes, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
