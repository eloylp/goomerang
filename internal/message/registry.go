package message

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type MessageRegistry map[string]proto.Message

func (r MessageRegistry) Register(key string, msg proto.Message) {
	r[key] = msg
}

func (r MessageRegistry) Message(key string) (proto.Message, error) {
	msg, ok := r[key]
	if !ok {
		return nil, fmt.Errorf("cannot found message with key: %s", key)
	}
	return msg, nil
}
