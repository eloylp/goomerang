package ws_test

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/goomerang"
	"go.eloylp.dev/goomerang/internal/ws"
)

func TestIsNotExpectedCloseError(t *testing.T) {
	cases := []struct {
		err         error
		nonExpected bool
	}{
		{websocket.ErrCloseSent, false},
		{&websocket.CloseError{
			Code: websocket.CloseNormalClosure,
		}, false},
		{&websocket.CloseError{
			Code: websocket.CloseAbnormalClosure,
		}, false},
		{nil, false},
		{&net.OpError{
			Op: "close",
		}, false},

		{errors.New("unknown error"), true},
		{&websocket.CloseError{
			Code: websocket.CloseMessageTooBig,
		}, true},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%T ---> %v", c.err, c.nonExpected), func(t *testing.T) {
			assert.Equal(t, c.nonExpected, ws.IsNotExpectedCloseError(c.err))
		})
	}
}

func TestMapErr(t *testing.T) {
	cases := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{"Internal ws lib errors are masked", websocket.ErrCloseSent, goomerang.ErrConnectionClosed},
		{"Unknown error is returned as is", errors.New("unknown error"), errors.New("ws error: unknown error")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.EqualError(t, ws.MapErr(c.err), c.expectedErr.Error())
		})
	}
}
