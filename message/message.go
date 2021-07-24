package message

import (
	"google.golang.org/protobuf/proto"
)

func FQDN(msg proto.Message) string {
	return string(msg.ProtoReflect().Descriptor().FullName())
}
