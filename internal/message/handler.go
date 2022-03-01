package message

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"go.eloylp.dev/goomerang"
)

type Handler interface {
	Handle(sender Sender, msg *goomerang.Message)
}

type HandlerFunc func(sender Sender, msg *goomerang.Message)

type Middleware func(h Handler) Handler

func (h HandlerFunc) Handle(sender Sender, msg *goomerang.Message) {
	h(sender, msg)
}

type Sender interface {
	Send(ctx context.Context, msg *goomerang.Message) error
}

type HandlerChainer struct {
	middlewares []Middleware
	handlers    map[string]Handler
	chains      map[string]Handler
	initiated   *int32
}

func NewHandlerChainer() *HandlerChainer {
	i := int32(0)
	return &HandlerChainer{
		handlers:  map[string]Handler{},
		chains:    map[string]Handler{},
		initiated: &i,
	}
}

func (hc *HandlerChainer) AppendHandler(chainName string, h Handler) {
	hc.mustNotBeInitiated()
	hc.handlers[chainName] = h
}

func (hc *HandlerChainer) mustNotBeInitiated() {
	if atomic.LoadInt32(hc.initiated) != 0 {
		panic(errors.New("handler chainer: handlers and middlewares can only be added before starting serving"))
	}
}

func (hc *HandlerChainer) AppendMiddleware(m Middleware) {
	hc.mustNotBeInitiated()
	hc.middlewares = append(hc.middlewares, m)
}

func (hc *HandlerChainer) PrepareChains() {
	atomic.StoreInt32(hc.initiated, 1)
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
