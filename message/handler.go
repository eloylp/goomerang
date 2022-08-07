package message

import (
	"go.eloylp.dev/goomerang/conn"
)

type Handler interface {
	Handle(sender Sender, msg *Message)
}

type HandlerFunc func(sender Sender, msg *Message)

type Middleware func(h Handler) Handler

func (h HandlerFunc) Handle(sender Sender, msg *Message) {
	h(sender, msg)
}

type Sender interface {
	Send(msg *Message) (payloadSize int, err error)
	ConnSlot() *conn.Slot
}
