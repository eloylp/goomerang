package engine

import (
	"fmt"
)

type AppendableRegistry map[string][]interface{}

func (r AppendableRegistry) Register(key string, elems ...interface{}) {
	hs, ok := r[key]
	if !ok {
		r[key] = elems
		return
	}
	hs = append(hs, elems...)
	r[key] = hs
}

func (r AppendableRegistry) Elems(key string) ([]interface{}, error) {
	h, ok := r[key]
	if !ok {
		return nil, fmt.Errorf("cannot found group of elems with key: %s", key)
	}
	return h, nil
}
