package message

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.eloylp.dev/goomerang/internal/message/protocol"
)

type FrameOption func(f *protocol.Frame)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

func FromFrame(frame *protocol.Frame, msgRegistry Registry) (*Request, error) {
	meta := &Metadata{
		Creation: frame.Creation.AsTime(),
		UUID:     frame.Uuid,
		Type:     frame.Type,
		IsRPC:    frame.IsRpc,
	}
	msg, err := msgRegistry.Message(frame.Type)
	if err != nil {
		return nil, fmt.Errorf("problems locating message: %w", err)
	}
	if err := proto.Unmarshal(frame.Payload, msg); err != nil {
		return nil, fmt.Errorf("parsing from message: %w", msg)
	}
	return &Request{
		Metadata: meta,
		Payload:  msg,
	}, nil
}

func Pack(msg proto.Message, opts ...FrameOption) ([]byte, error) {
	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	frame := &protocol.Frame{
		Type:     FQDN(msg),
		Payload:  payload,
		Creation: timestamppb.New(time.Now()),
	}
	for _, opt := range opts {
		opt(frame)
	}
	bytes, err := proto.Marshal(frame)
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

func FrameWithUUID(uuid string) FrameOption {
	return func(f *protocol.Frame) {
		f.Uuid = uuid
	}
}

func FrameIsRPC() FrameOption {
	return func(f *protocol.Frame) {
		f.IsRpc = true
	}
}
