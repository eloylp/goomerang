package server

import (
	"context"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
)

type immediateSender struct {
	s        *Server
	connSlot connSlot
}

func (so *immediateSender) Send(ctx context.Context, msg *message.Message) (int, error) {
	ch := make(chan sendResponse, 1)
	go func() {
		defer close(ch)
		payloadSize, m, err := messaging.Pack(msg)
		if err != nil {
			ch <- sendResponse{payloadSize, err}
			return
		}
		ch <- sendResponse{payloadSize, so.connSlot.write(m)}
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case resp := <-ch:
		return resp.payloadSize, resp.err
	}
}

type sendResponse struct {
	payloadSize int
	err         error
}

type bufferedSender struct {
	reply *message.Message
}

func (so *bufferedSender) Send(_ context.Context, msg *message.Message) (int, error) {
	so.reply = msg
	return 0, nil
}

func (so *bufferedSender) Reply() *message.Message {
	return so.reply
}
