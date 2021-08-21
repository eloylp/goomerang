package server

import (
	"errors"
	"fmt"
)

var (
	ErrClientDisconnected = errors.New("client properly closed connection")
)

type HandlerError struct {
	Message string
	Code    int32
}

func NewHandlerErrorWith(message string, code int32) *HandlerError {
	return &HandlerError{Message: message, Code: code}
}

func NewHandlerError(message string) *HandlerError {
	return &HandlerError{Message: message, Code: 0}
}

func (h HandlerError) Error() string {
	return fmt.Sprintf("handler error code %v: %s", h.Code, h.Message)
}
