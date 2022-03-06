package server

import (
	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/message"
)

func doRPC(handler message.Handler, cs connSlot, msg *goomerang.Message) error {
	ops := &bufferedSender{}
	handler.Handle(ops, msg)
	responseMsg, err := message.Pack(ops.Reply(), message.FrameWithUUID(msg.Metadata.UUID), message.FrameIsRPC())
	if err != nil {
		return err
	}
	if err := cs.write(responseMsg); err != nil {
		return err
	}
	return nil
}
