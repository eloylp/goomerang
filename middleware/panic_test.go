//go:build unit

package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/middleware"
)

func TestPanic(t *testing.T) {
	arbiter := test.NewArbiter(t)
	m := middleware.Panic(func(p interface{}) {
		arbiter.ItsAFactThat("PANIC_HANDLER_CALLED")
		require.NotNil(t, p)
	})
	h := message.HandlerFunc(func(_ message.Sender, _ *message.Message) {
		panic("panic!")
	})
	m(h).Handle(nil, nil)
	arbiter.RequireHappened("PANIC_HANDLER_CALLED")
}
