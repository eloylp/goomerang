package client

import (
	"errors"
)

var (
	ErrNotRunning     = errors.New("not running")
	ErrAlreadyRunning = errors.New("already running")
)
