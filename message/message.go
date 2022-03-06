package message

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	Metadata *Metadata
	Payload  proto.Message
	Header   Header
}

type Metadata struct {
	Creation time.Time
	UUID     string
	Type     string
	IsRPC    bool
}

type Header map[string]string

func (h Header) Get(key string) string {
	return h[key]
}

func (h Header) Add(key, value string) {
	h[key] = value
}
