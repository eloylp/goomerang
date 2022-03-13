package messaging

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}

func FromFrame(frame *protocol.Frame, msgRegistry Registry) (*message.Message, error) {
	meta := &message.Metadata{
		Creation: frame.Creation.AsTime(),
		UUID:     frame.Uuid,
		Type:     frame.Type,
		IsSync:   frame.IsSync,
	}
	msg, err := msgRegistry.Message(frame.Type)
	if err != nil {
		return nil, fmt.Errorf("problems locating message: %w", err)
	}
	if err := proto.Unmarshal(frame.Payload, msg); err != nil {
		return nil, fmt.Errorf("parsing from message: %s", FQDN(msg))
	}
	return &message.Message{
		Metadata: meta,
		Payload:  msg,
		Header:   message.Header(frame.Headers),
	}, nil
}

func Pack(msg *message.Message, opts ...FrameOption) ([]byte, error) {
	payload, err := proto.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	frame := &protocol.Frame{
		Type:     FQDN(msg.Payload),
		Payload:  payload,
		Creation: timestamppb.New(time.Now()),
		Headers:  msg.Header,
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
