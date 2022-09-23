//go:build integration

package goomerang_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
)

func TestClientReturnsKnownErrOnConnFailure(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)

	s, run := Server(t)
	s.Handle(defaultMsg.Payload, echoHandler)
	run()
	defer s.Shutdown(defaultCtx)

	goomerangProxy, err := proxyClient.CreateProxy(t.Name(), "", s.Addr())
	require.NoError(t, err)
	defer goomerangProxy.Delete()

	c, connect := Client(t,
		client.WithServerAddr(goomerangProxy.Listen),
		client.WithOnCloseHook(func() {
			arbiter.ItsAFactThat("CLIENT_ONCLOSE_HOOK")
		}),
		client.WithOnStatusChangeHook(statusChangesHook(arbiter, "client")),
	)
	connect()
	defer c.Close(defaultCtx)

	require.NoError(t, goomerangProxy.Disable())

	require.Eventually(t, func() bool {
		_, err := c.Send(defaultMsg)
		return errors.Is(err, client.ErrNotRunning)
	}, time.Second, time.Millisecond, "it was expected client to return ErrNotRunning")

	require.Eventually(t, func() bool {
		_, _, err = c.SendSync(defaultCtx, defaultMsg)
		return errors.Is(err, client.ErrNotRunning)
	}, time.Second, time.Millisecond, "it was expected client to return ErrNotRunning")

	arbiter.RequireHappened("CLIENT_ONCLOSE_HOOK")
	arbiter.RequireHappened("CLIENT_WAS_CLOSING")
	arbiter.RequireHappened("CLIENT_WAS_CLOSED")
}
