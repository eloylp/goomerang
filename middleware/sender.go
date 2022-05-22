package middleware

import (
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

func (s *Sender) Send(msg *message.Message) (int, error) {
	w, err := s.sender.Send(msg)
	s.msg = msg
	s.bytes = w
	return w, err
}
