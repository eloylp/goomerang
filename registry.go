package goomerang

import (
	"fmt"
)

type Registry map[string]ServerHandler

func (r Registry) Register(key string, h ServerHandler) {
	r[key] = h
}

func (r Registry) Handler(key string) (ServerHandler, error) {
	h, ok := r[key]
	if !ok {
		return nil, fmt.Errorf("cannot found handler with key: %s", key)
	}
	return h, nil
}
