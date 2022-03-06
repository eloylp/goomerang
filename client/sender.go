package client

import (
	"context"

	"go.eloylp.dev/goomerang/message"
)

type immediateSender struct {
	c *Client
}

func (co *immediateSender) Send(ctx context.Context, msg *message.Message) error {
	return co.c.Send(ctx, msg)
}
