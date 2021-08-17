package message

import (
	"fmt"
)

type Registry map[string][]interface{}

func (r Registry) Register(key string, handlers ...interface{}) {
	hs, ok := r[key]
	if !ok {
		r[key] = handlers
		return
	}
	hs = append(hs, handlers...)
	r[key] = hs
}

func (r Registry) Handler(key string) ([]interface{}, error) {
	h, ok := r[key]
	if !ok {
		return nil, fmt.Errorf("cannot found handler with key: %s", key)
	}
	return h, nil
}
