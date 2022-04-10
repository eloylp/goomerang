package middleware

import (
	"context"

	"go.eloylp.dev/goomerang/message"
)

type Sender struct {
	bytes  int
	msg    *message.Message
	sender message.Sender
}

func NewSender(sender message.Sender) *Sender {
	return &Sender{sender: sender}
}

func (s *Sender) Msg() *message.Message {
	return s.msg
}

func (s *Sender) Bytes() int {
	return s.bytes
}

func (s *Sender) Send(ctx context.Context, msg *message.Message) (int, error) {
	w, err := s.sender.Send(ctx, msg)
	s.bytes = w
	return w, err
}
