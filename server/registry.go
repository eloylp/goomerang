package server

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Registry map[string][2]interface{}

func (r Registry) Register(msg proto.Message, h ServerHandler) {
	key := msg.ProtoReflect().Descriptor().FullName()
	r[string(key)] = [2]interface{}{msg, h}
}

func (r Registry) Handler(key string) (proto.Message, ServerHandler, error) {
	slot, ok := r[key]
	if !ok {
		return nil, nil, fmt.Errorf("cannot found handler with key: %s", key)
	}
	return slot[0].(proto.Message), slot[1].(ServerHandler), nil
}
