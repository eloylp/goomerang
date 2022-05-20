package server

import (
	"errors"
)

var (
	ErrClosing        = errors.New("server is closing")
	ErrAlreadyRunning = errors.New("server is already running")
	ErrNotRunning     = errors.New("server is not running")
)
