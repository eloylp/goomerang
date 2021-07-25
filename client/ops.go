package client

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Ops interface {
	Send(ctx context.Context, msg proto.Message) error
}

type clientOps struct {
	c *Client
}

func (co *clientOps) Send(ctx context.Context, msg proto.Message) error {
	return co.c.Send(ctx, msg)
}
