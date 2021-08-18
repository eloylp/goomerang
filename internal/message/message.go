package message

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message/protocol"
)

type FrameOption func(f *protocol.Frame)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

func Pack(msg proto.Message, opts ...FrameOption) ([]byte, error) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	envelope := &protocol.Frame{
		Type:    FQDN(msg),
		Payload: payload,
	}
	for _, opt := range opts {
		opt(envelope)
	}
	bytes, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func UnPack(data []byte) (*protocol.Frame, error) {
	frame := &protocol.Frame{}
	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, fmt.Errorf("protocol decoding err: %w", err)
	}
	return frame, nil
}

func FrameWithUuid(uuid string) FrameOption {
	return func(f *protocol.Frame) {
		f.Uuid = uuid
	}
}
