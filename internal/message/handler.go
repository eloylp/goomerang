package message

import (
	"context"
	"fmt"
)

type Handler interface {
	Handle(sender Sender, msg *Message)
}

type HandlerFunc func(sender Sender, msg *Message)

type Middleware func(h Handler) Handler

func (h HandlerFunc) Handle(sender Sender, msg *Message) {
	h(sender, msg)
}

type Sender interface {
	Send(ctx context.Context, msg *Message) error
}

type HandlerChainer struct {
	middlewares []Middleware
	handlers    map[string]Handler
	chains      map[string]Handler
}

func NewHandlerChainer() *HandlerChainer {
	return &HandlerChainer{
		handlers: map[string]Handler{},
		chains:   map[string]Handler{},
	}
}

func (hc *HandlerChainer) AppendHandler(chainName string, h Handler) {
	hc.handlers[chainName] = h
}

func (hc *HandlerChainer) AppendMiddleware(m Middleware) {
	hc.middlewares = append(hc.middlewares, m)
}

func (hc *HandlerChainer) PrepareChains() {
	for key, h := range hc.handlers {
		hc.chains[key] = hc.middlewareFor(h, hc.middlewares...)
	}
}
func (hc *HandlerChainer) middlewareFor(h Handler, m ...Middleware) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func (hc *HandlerChainer) Handler(chain string) (Handler, error) {
	c, ok := hc.chains[chain]
	if !ok {
		return nil, fmt.Errorf("handler chainer: chain %q not found", chain)
	}
	return c, nil
}
