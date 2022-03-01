package message

import (
	"go.eloylp.dev/goomerang/internal/message/protocol"
)

type FrameOption func(f *protocol.Frame)

func FrameWithUUID(uuid string) FrameOption {
	return func(f *protocol.Frame) {
		f.Uuid = uuid
	}
}

func FrameIsRPC() FrameOption {
	return func(f *protocol.Frame) {
		f.IsRpc = true
	}
}
