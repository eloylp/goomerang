package message_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/internal/test"
)

func TestMiddlewareRegistry(t *testing.T) {
	arbiter := test.NewArbiter(t)
	mr := message.NewHandlerChainer()

	handler := message.HandlerFunc(func(sender message.Sender, msg *message.Request) {
		arbiter.ItsAFactThat("HANDLER_EXECUTED")
		sender.Send(context.Background(), &testMessages.PingPong{})
	})

	middleware := message.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(sender message.Sender, msg *message.Request) {
			arbiter.ItsAFactThat("MIDDLEWARE_EXECUTED")
			h.Handle(sender, msg)
		})
	})

	mr.AppendHandler("chain1", handler)
	mr.AppendMiddleware(middleware)
	mr.PrepareChains()

	h, err := mr.Handler("chain1")
	require.NoError(t, err)

	h.Handle(&FakeSender{a: arbiter}, &message.Request{})

	arbiter.RequireHappenedInOrder("MIDDLEWARE_EXECUTED", "HANDLER_EXECUTED")
	arbiter.RequireHappenedInOrder("HANDLER_EXECUTED", "SENDER_CALLED")
}

func TestMiddlewareRegistryNotFound(t *testing.T) {
	mr := message.NewHandlerChainer()
	mr.PrepareChains()
	_, err := mr.Handler("NON_EXISTENT")
	assert.EqualError(t, err, `handler chainer: chain "NON_EXISTENT" not found`)
}

type FakeSender struct {
	a *test.Arbiter
}

func (f *FakeSender) Send(ctx context.Context, msg proto.Message) error {
	f.a.ItsAFactThat("SENDER_CALLED")
	return nil
}
