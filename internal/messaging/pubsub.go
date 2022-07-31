package messaging

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/messaging/protocol"
	"go.eloylp.dev/goomerang/message"
)

const userMessageHeadersPrefix = "message-"

func MessageForPublish(topic string, msg *message.Message) (*message.Message, error) {
	data, err := proto.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("messageForPublish: %v", err)
	}
	pubMsg := message.New().SetPayload(&protocol.PublishCmd{
		Topic:   topic,
		Kind:    msg.Metadata.Kind,
		Message: data,
	})
	pubMsg.Header = packHeaders(msg.Header)
	return pubMsg, nil
}

func packHeaders(headers message.Header) message.Header {
	packedHeaders := message.Header{}
	for k, v := range headers {
		packedHeaders.Set(fmt.Sprintf("%s%s", userMessageHeadersPrefix, k), v)
	}
	return packedHeaders
}

func MessageFromPublish(registry message.Registry, msg *message.Message) (*message.Message, error) {
	pubCmd, ok := msg.Payload.(*protocol.PublishCmd)
	if !ok {
		return nil, fmt.Errorf("messageFromPublish: %v", "input message its not a protocol.PublishCommand one")
	}
	clientMsg, err := registry.Message(pubCmd.Kind)
	if err != nil {
		return nil, fmt.Errorf("messageFromPublish: %v", err)
	}
	if err := proto.Unmarshal(pubCmd.Message, clientMsg); err != nil {
		return nil, fmt.Errorf("messageFromPublish: %v", err)
	}
	pubMsg := message.New().SetPayload(clientMsg)
	pubMsg.Header = unpackHeaders(msg.Header)
	return pubMsg, nil
}

func unpackHeaders(headers message.Header) message.Header {
	unpackedHeaders := message.Header{}
	for k, v := range headers {
		if strings.HasPrefix(k, userMessageHeadersPrefix) {
			key := strings.TrimPrefix(k, userMessageHeadersPrefix)
			unpackedHeaders.Set(key, v)
		}
	}
	return unpackedHeaders
}
