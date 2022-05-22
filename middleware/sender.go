package middleware

import (
	"time"

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

type MeteredSender struct {
	bytes      int
	msg        *message.Message
	sender     message.Sender
	promConfig PromConfig
}

func NewMeteredSender(sender message.Sender, promConfig PromConfig) *MeteredSender {
	return &MeteredSender{
		sender:     sender,
		promConfig: promConfig,
	}
}

func (s *MeteredSender) Msg() *message.Message {
	return s.msg
}

func (s *MeteredSender) Bytes() int {
	return s.bytes
}

func (s *MeteredSender) Send(msg *message.Message) (int, error) {
	start := time.Now()
	w, err := s.sender.Send(msg)
	if err == nil {
		s.promConfig.MessageSentTime.WithLabelValues(msg.Metadata.Type).Observe(time.Since(start).Seconds())
	}
	s.msg = msg
	s.bytes = w
	return w, err
}
