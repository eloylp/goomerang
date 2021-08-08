package server

import (
	"context"
	"go.eloylp.dev/goomerang/internal/message"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
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
	m, err := message.PrepareMessage(msg)
	if err != nil {
		return err
	}
	return so.c.WriteMessage(websocket.BinaryMessage, m)
}

func (so *serverOpts) Shutdown(ctx context.Context) error {
	return so.s.Shutdown(ctx)
}
