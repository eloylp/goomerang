package server

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Ops interface {
	Send(ctx context.Context, msg proto.Message) error
	Shutdown(ctx context.Context) error
}

type serverOpts struct {
	s *Server
}

func (so *serverOpts) Send(ctx context.Context, msg proto.Message) error {
	return so.s.Send(ctx, msg)
}

func (so *serverOpts) Shutdown(ctx context.Context) error {
	return so.s.Shutdown(ctx)
}
