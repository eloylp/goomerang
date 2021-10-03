package message

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

type Registry map[string]proto.Message

func (r Registry) Register(key string, msg proto.Message) {
	r[key] = msg
}

func (r Registry) Message(key string) (proto.Message, error) {
	msg, ok := r[key]
	if !ok {
		return nil, fmt.Errorf("cannot found message with key: %s", key)
	}
	msgCopy := reflect.New(reflect.ValueOf(msg).Elem().Type()).Interface().(proto.Message)
	return msgCopy, nil
}
