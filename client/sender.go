package client

import (
	"context"

	"go.eloylp.dev/goomerang"
)

type immediateSender struct {
	c *Client
}

func (co *immediateSender) Send(ctx context.Context, msg *goomerang.Message) error {
	return co.c.Send(ctx, msg)
}
