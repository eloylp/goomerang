package server

import (
	"context"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/message"
)

type immediateSender struct {
	s        *Server
	connSlot connSlot
}

func (so *immediateSender) Send(ctx context.Context, msg *goomerang.Message) error {
	m, err := message.Pack(msg)
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
	reply *goomerang.Message
}

func (so *bufferedSender) Send(_ context.Context, msg *goomerang.Message) error {
	so.reply = msg
	return nil
}

func (so *bufferedSender) Reply() *goomerang.Message {
	return so.reply
}
