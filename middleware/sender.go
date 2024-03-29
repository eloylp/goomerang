package middleware

import (
	"time"

	"go.eloylp.dev/goomerang/conn"
	"go.eloylp.dev/goomerang/message"
)

// Sender is just a handy middleware for gathering
// the message and the bytes sent after the operation.
type Sender struct {
	bytes  int
	msg    *message.Message
	sender message.Sender
}

func (s *Sender) ConnSlot() *conn.Slot {
	return s.sender.ConnSlot()
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

// MeteredSender is an instrumented sender. It's very handy
// to know the sent message,size and time. It requires
// the PromConfig parameter, which should be populated beforehand
// with the needed histograms.
type MeteredSender struct {
	bytes      int
	msg        *message.Message
	sender     message.Sender
	promConfig PromConfig
}

func (s *MeteredSender) ConnSlot() *conn.Slot {
	return s.sender.ConnSlot()
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
		s.promConfig.MessageSentTime.WithLabelValues(msg.Metadata.Kind).Observe(time.Since(start).Seconds())
	}
	s.msg = msg
	s.bytes = w
	return w, err
}
