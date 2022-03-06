package server

import (
	"context"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
)

type immediateSender struct {
	s        *Server
	connSlot connSlot
}

func (so *immediateSender) Send(ctx context.Context, msg *message.Message) error {
	m, err := messaging.Pack(msg)
	if err != nil {
		return err
	}
	ch := make(chan error, 1)
	go func() {
		so.connSlot.l.Lock()
		defer so.connSlot.l.Unlock()
		ch <- so.connSlot.c.WriteMessage(websocket.BinaryMessage, m)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}

type bufferedSender struct {
	reply *message.Message
}

func (so *bufferedSender) Send(_ context.Context, msg *message.Message) error {
	so.reply = msg
	return nil
}

func (so *bufferedSender) Reply() *message.Message {
	return so.reply
}
