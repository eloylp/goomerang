package messaging

import (
	"go.eloylp.dev/goomerang/internal/messaging/protocol"
)

type FrameOption func(f *protocol.Frame)

func FrameWithUUID(uuid string) FrameOption {
	return func(f *protocol.Frame) {
		f.Uuid = uuid
	}
}

func FrameIsSync() FrameOption {
	return func(f *protocol.Frame) {
		f.IsSync = true
	}
}
