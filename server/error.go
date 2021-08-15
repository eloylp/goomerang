package server

import (
	"errors"
)

var (
	ErrClientDisconnected = errors.New("client properly closed connection")
)
