package server

import (
	"context"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
)

type Ops interface {
	Send(ctx context.Context, msg proto.Message) error
	Shutdown(ctx context.Context) error
}

type serverOpts struct {
	s *Server
	c *websocket.Conn
}

func (so *serverOpts) Send(ctx context.Context, msg proto.Message) error {
	m, err := message.Pack(msg)
	if err != nil {
		return err
	}
	return so.c.WriteMessage(websocket.BinaryMessage, m)
}

func (so *serverOpts) Shutdown(ctx context.Context) error {
	return so.s.Shutdown(ctx)
}

type serverOptsReqRep struct {
	replies []proto.Message
}

func (so *serverOptsReqRep) Send(ctx context.Context, msg proto.Message) error {
	so.replies = append(so.replies, msg)
	return nil
}

func (so *serverOptsReqRep) Shutdown(ctx context.Context) error {
	return nil
}

func (so *serverOptsReqRep) Index(i int) proto.Message {
	return so.replies[i]
}
