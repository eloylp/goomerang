package goomerang

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type PeerOps interface {
	Send(ctx context.Context, msg proto.Message) error
}
