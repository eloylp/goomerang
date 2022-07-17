package server

import (
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/ws"
)

type stdSender struct {
	connSlot *connSlot
	status   func() uint32
}

func (s *stdSender) Send(msg *message.Message) (int, error) {
	if s.status() != ws.StatusRunning {
		return 0, ErrNotRunning
	}
	payloadSize, m, err := messaging.Pack(msg)
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.connSlot.write(m)
}

type SyncSender struct {
	cs          *connSlot
	status      func() uint32
	prevMsgUUID string
}

func (s *SyncSender) Send(msg *message.Message) (int, error) {
	if s.status() != ws.StatusRunning {
		return 0, ErrNotRunning
	}
	payloadSize, m, err := messaging.Pack(msg, messaging.FrameWithUUID(s.prevMsgUUID), messaging.FrameIsSync())
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.cs.write(m)
}
