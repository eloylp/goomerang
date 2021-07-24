package server

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
)

type Registry map[string]slot

type slot struct {
	msg      proto.Message
	handlers []ServerHandler
}

func (r Registry) Register(msg proto.Message, h ServerHandler) {
	key := message.FQDN(msg)
	r[key] = slot{
		msg:      msg,
		handlers: []ServerHandler{h},
	}
}

func (r Registry) Handler(key string) (proto.Message, []ServerHandler, error) {
	slot, ok := r[key]
	if !ok {
		return nil, nil, fmt.Errorf("cannot found handler with key: %s", key)
	}
	return slot.msg, slot.handlers, nil
}
