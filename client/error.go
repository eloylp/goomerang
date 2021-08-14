package client

import (
	"errors"
)

var (
	ErrServerDisconnected = errors.New("server properly closed connection")
)
