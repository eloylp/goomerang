package rpc

import (
	"fmt"
)

type HandlerError struct {
	Message string
	Code    int32
}

func NewHandlerError(message string, code int32) *HandlerError {
	return &HandlerError{Message: message, Code: code}
}

func (h HandlerError) Error() string {
	return fmt.Sprintf("handler error code %v: %s", h.Code, h.Message)
}
