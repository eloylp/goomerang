package server

import (
	"context"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
)

type Sender interface {
	Send(ctx context.Context, msg proto.Message) error
}

type immediateSender struct {
	s        *Server
	connSlot connSlot
}

func (so *immediateSender) Send(ctx context.Context, msg proto.Message) error {
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
	reply proto.Message
}

func (so *bufferedSender) Send(_ context.Context, msg proto.Message) error {
	so.reply = msg
	return nil
}

func (so *bufferedSender) Reply() proto.Message {
	return so.reply
}
