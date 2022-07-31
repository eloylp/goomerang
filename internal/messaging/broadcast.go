package messaging

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
)

func MessageForBroadcast(msg *message.Message) (*message.Message, error) {
	data, err := proto.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("messageForBroadcast: %v", err)
	}
	brMsg := message.New().SetPayload(&protocol.BroadcastCmd{
		Kind:    msg.Metadata.Kind,
		Message: data,
	})
	brMsg.Header = packHeaders(msg.Header)
	return brMsg, nil
}

func MessageFromBroadcast(registry message.Registry, msg *message.Message) (*message.Message, error) {
	brCmd, ok := msg.Payload.(*protocol.BroadcastCmd)
	if !ok {
		return nil, fmt.Errorf("messageFromBroadcast: %v", "input message its not a protocol.PublishBroadcast one")
	}
	clientMsg, err := registry.Message(brCmd.Kind)
	if err != nil {
		return nil, fmt.Errorf("messageFromBroadcast: %v", err)
	}
	if err := proto.Unmarshal(brCmd.Message, clientMsg); err != nil {
		return nil, fmt.Errorf("messageFromBroadcast: %v", err)
	}
	brMsg := message.New().SetPayload(clientMsg)
	brMsg.Header = unpackHeaders(msg.Header)
	return brMsg, nil
}
