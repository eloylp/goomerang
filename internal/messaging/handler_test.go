package messaging_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/messaging"
	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
)

func TestMiddlewareRegistry(t *testing.T) {
	arbiter := test.NewArbiter(t)
	mr := messaging.NewHandlerChainer()

	handler := message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("HANDLER_EXECUTED")
		resp := &message.Message{
			Payload: &testMessages.MessageV1{},
		}
		sender.Send(resp)
	})

	middleware1 := message.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("MIDDLEWARE_1_EXECUTED")
			h.Handle(sender, msg)
		})
	})

	middleware2 := message.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("MIDDLEWARE_2_EXECUTED")
			h.Handle(sender, msg)
		})
	})

	mr.AppendHandler("chain1", handler)
	mr.AppendMiddleware(middleware1)
	mr.AppendMiddleware(middleware2)
	mr.PrepareChains()

	h, err := mr.Handler("chain1")
	require.NoError(t, err)

	h.Handle(&FakeSender{a: arbiter}, &message.Message{})

	arbiter.RequireHappenedInOrder("MIDDLEWARE_1_EXECUTED", "MIDDLEWARE_2_EXECUTED")
	arbiter.RequireHappenedInOrder("MIDDLEWARE_2_EXECUTED", "HANDLER_EXECUTED")
	arbiter.RequireHappenedInOrder("HANDLER_EXECUTED", "SENDER_CALLED")
}

func TestMiddlewareRegistryNotFound(t *testing.T) {
	mr := messaging.NewHandlerChainer()
	mr.PrepareChains()
	_, err := mr.Handler("NON_EXISTENT")
	assert.EqualError(t, err, `handler chainer: chain "NON_EXISTENT" not found`)
}

type FakeSender struct {
	a *test.Arbiter
}

func (f *FakeSender) Send(msg *message.Message) (int, error) {
	f.a.ItsAFactThat("SENDER_CALLED")
	return 0, nil
}

func TestHandlerChainer_PanicsAfterPrepareChains(t *testing.T) {
	hc := messaging.NewHandlerChainer()
	hc.PrepareChains()
	expectedError := "handler chainer: handlers and middlewares can only be added before starting serving"
	require.PanicsWithError(t, expectedError, func() {
		hc.AppendHandler("chain", message.HandlerFunc(func(sender message.Sender, msg *message.Message) {

		}))
	}, "should be panicked when registering a handler after PrepareChains()")

	require.PanicsWithError(t, expectedError, func() {
		hc.AppendMiddleware(func(h message.Handler) message.Handler {
			return message.HandlerFunc(func(sender message.Sender, msg *message.Message) {})
		})
	}, func() {

	}, "should be panicked when registering a middleware after PrepareChains()")
}
