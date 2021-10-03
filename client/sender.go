package client

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type Sender interface {
	Send(ctx context.Context, msg proto.Message) error
}

type immediateSender struct {
	c *Client
}

func (co *immediateSender) Send(ctx context.Context, msg proto.Message) error {
	return co.c.Send(ctx, msg)
}
