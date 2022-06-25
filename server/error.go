package server

import (
	"errors"
)

var (
	ErrClosing        = errors.New("server is closing")
	ErrClosed         = errors.New("server is closed")
	ErrAlreadyRunning = errors.New("server is already running")
	ErrNotRunning     = errors.New("server is not running")
)
