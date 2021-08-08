package message

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message/protocol"
)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

func Pack(msg proto.Message) ([]byte, error) {
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

func UnPack(registry Registry, data []byte) (proto.Message, []interface{}, error) {
	frame := &protocol.Frame{}
	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, nil, fmt.Errorf("protocol decoding err: %w", err)
	}
	msg, handlers, err := registry.Handler(frame.GetType())
	if err != nil {
		return nil, nil, fmt.Errorf("handler registry: %w", err)
	}
	if err = proto.Unmarshal(frame.Payload, msg); err != nil {
		return nil, nil, fmt.Errorf("payload decoding err: %w", err)
	}
	return msg, handlers, nil
}
