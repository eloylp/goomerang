package client

import (
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/server"
)

type MultiReply struct {
	Replies []*Reply
}

func (m *MultiReply) First() *Reply {
	return m.Replies[0]
}

func (m *MultiReply) Index(i int) *Reply {
	return m.Replies[i]
}

type Reply struct {
	Message proto.Message
	Err     *server.HandlerError
}
