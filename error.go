package goomerang

import (
	"errors"
)

var (
	ErrConnectionClosed = errors.New("tried to operate on an already closed connection")
)
