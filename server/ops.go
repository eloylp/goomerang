package server

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
)

type Ops interface {
	Send(ctx context.Context, msg proto.Message) error
}

type serverOpts struct {
	s *Server
	c *websocket.Conn
	l *sync.Mutex
}

func (so *serverOpts) Send(ctx context.Context, msg proto.Message) error {
	m, err := message.Pack(msg)
	if err != nil {
		return err
	}
	so.l.Lock()
	defer so.l.Unlock()
	return so.c.WriteMessage(websocket.BinaryMessage, m)
}

type serverOptsReqRep struct {
	replies []proto.Message
}

func (so *serverOptsReqRep) Send(ctx context.Context, msg proto.Message) error {
	so.replies = append(so.replies, msg)
	return nil
}

func (so *serverOptsReqRep) Index(i int) proto.Message {
	return so.replies[i]
}
