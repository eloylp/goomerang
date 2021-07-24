package server

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
)

type Registry map[string]slot

type slot struct {
	msg      proto.Message
	handlers []Handler
}

func (r Registry) Register(msg proto.Message, handlers ...Handler) {
	key := message.FQDN(msg)
	s, ok := r[key]
	if !ok {
		r[key] = slot{
			msg:      msg,
			handlers: handlers,
		}
		return
	}
	s.handlers = append(s.handlers, handlers...)
	r[key] = s
}

func (r Registry) Handler(key string) (proto.Message, []Handler, error) {
	slot, ok := r[key]
	if !ok {
		return nil, nil, fmt.Errorf("cannot found handler with key: %s", key)
	}
	return slot.msg, slot.handlers, nil
}
