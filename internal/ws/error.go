package ws

import (
	"errors"
	"fmt"
	"net"

	"github.com/gorilla/websocket"

	"go.eloylp.dev/goomerang"
)

func MapErr(err error) error {
	if errors.Is(err, websocket.ErrCloseSent) {
		return goomerang.ErrConnectionClosed
	}
	return fmt.Errorf("ws error: %v", err)
}

// IsNotExpectedCloseError should be only used in shutdown sequences,
// in which we can expect some connection errors.
func IsNotExpectedCloseError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, websocket.ErrCloseSent) {
		return false
	}
	if err, ok := err.(*net.OpError); ok {
		return err.Op != "close"
	}
	if _, ok := err.(*websocket.CloseError); !ok {
		return true
	}
	return websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure)
}
