package server

import (
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
)

type stdSender struct {
	connSlot connSlot
}

func (s *stdSender) Send(msg *message.Message) (int, error) {
	payloadSize, m, err := messaging.Pack(msg)
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.connSlot.write(m)
}

type SyncSender struct {
	cs          connSlot
	prevMsgUUID string
}

func (s *SyncSender) Send(msg *message.Message) (int, error) {
	payloadSize, m, err := messaging.Pack(msg, messaging.FrameWithUUID(s.prevMsgUUID), messaging.FrameIsSync())
	if err != nil {
		return payloadSize, err
	}
	return payloadSize, s.cs.write(m)

}
