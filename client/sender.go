package client

import (
	"context"
	"go.eloylp.dev/goomerang/internal/message"

	"google.golang.org/protobuf/proto"
)

type Sender interface {
	Send(ctx context.Context, msg proto.Message) error
}

type immediateSender struct {
	c *Client
}

func (co *immediateSender) Send(ctx context.Context, msg *message.Message) error {
	return co.c.Send(ctx, msg)
}
