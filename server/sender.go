package server

import (
	"go.eloylp.dev/goomerang/conn"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/ws"
)

type stdSender struct {
	connSlot *conn.Slot
	status   func() uint32
}

func (s *stdSender) ConnSlot() *conn.Slot {
	return s.connSlot
}

func (s *stdSender) Send(msg *message.Message) (int, error) {
	if s.status() != ws.StatusRunning {
		return 0, ErrNotRunning
	}
	payloadSize, m, err := messaging.Pack(msg)
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.connSlot.Write(m)
}

type SyncSender struct {
	cs          *conn.Slot
	status      func() uint32
	prevMsgUUID string
}

func (s *SyncSender) ConnSlot() *conn.Slot {
	return s.cs
}

func (s *SyncSender) Send(msg *message.Message) (int, error) {
	if s.status() != ws.StatusRunning {
		return 0, ErrNotRunning
	}
	payloadSize, m, err := messaging.Pack(msg, messaging.FrameWithUUID(s.prevMsgUUID), messaging.FrameIsSync())
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.cs.Write(m)
}
