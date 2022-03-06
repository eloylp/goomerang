package server

import (
	"go.eloylp.dev/goomerang/message"

	"go.eloylp.dev/goomerang/internal/messaging"
)

func doRPC(handler message.Handler, cs connSlot, msg *message.Message) error {
	ops := &bufferedSender{}
	handler.Handle(ops, msg)
	responseMsg, err := messaging.Pack(ops.Reply(), messaging.FrameWithUUID(msg.Metadata.UUID), messaging.FrameIsRPC())
	if err != nil {
		return err
	}
	if err := cs.write(responseMsg); err != nil {
		return err
	}
	return nil
}
