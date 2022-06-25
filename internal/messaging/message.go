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
		Creation:    frame.Creation.AsTime(),
		UUID:        frame.Uuid,
		Kind:        frame.Kind,
		PayloadSize: int(frame.PayloadSize),
		IsSync:      frame.IsSync,
	}
	msg, err := msgRegistry.Message(frame.Kind)
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

func Pack(msg *message.Message, opts ...FrameOption) (int, []byte, error) {
	payload, err := proto.Marshal(msg.Payload)
	if err != nil {
		return 0, nil, err
	}
	payloadSize := len(payload)
	frame := &protocol.Frame{
		Kind:        FQDN(msg.Payload),
		PayloadSize: int64(payloadSize),
		Payload:     payload,
		Creation:    timestamppb.New(time.Now()),
		Headers:     msg.Header,
	}
	for _, opt := range opts {
		opt(frame)
	}
	bytes, err := proto.Marshal(frame)
	if err != nil {
		return payloadSize, nil, err
	}
	return payloadSize, bytes, nil
}

func UnPack(data []byte) (*protocol.Frame, error) {
	frame := &protocol.Frame{}
	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, fmt.Errorf("protocol decoding err: %w", err)
	}
	return frame, nil
}
