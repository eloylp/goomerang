package server

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
)

type Sender interface {
	Send(ctx context.Context, msg proto.Message) error
}

type immediateSender struct {
	s *Server
	c *websocket.Conn
	l *sync.Mutex
}

func (so *immediateSender) Send(ctx context.Context, msg proto.Message) error {
	m, err := message.Pack(msg)
	if err != nil {
		return err
	}
	so.l.Lock()
	defer so.l.Unlock()
	return so.c.WriteMessage(websocket.BinaryMessage, m)
}

type bufferedSender struct {
	replies []proto.Message
}

func (so *bufferedSender) Send(_ context.Context, msg proto.Message) error {
	so.replies = append(so.replies, msg)
	return nil
}

func (so *bufferedSender) Index(i int) proto.Message {
	return so.replies[i]
}
