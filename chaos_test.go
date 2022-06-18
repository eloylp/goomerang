package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
)

func TestClientReturnsKnownErrOnConnFailure(t *testing.T) {
	arbiter := test.NewArbiter(t)

	s, run := PrepareServer(t)
	s.Handle(defaultMsg.Payload, echoHandler)
	run()
	defer s.Shutdown(defaultCtx)

	goomerangProxy, err := proxyClient.CreateProxy(t.Name(), "", s.Addr())
	require.NoError(t, err)
	defer goomerangProxy.Delete()

	c, connect := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithServerAddr(goomerangProxy.Listen),
		client.WithOnCloseHook(func() {
			arbiter.ItsAFactThat("CLIENT_ONCLOSE_HOOK")
		}),
		client.WithOnStatusChangeHook(statusChangesHook(arbiter, "client")),
	)
	connect()
	defer c.Close(defaultCtx)

	require.NoError(t, goomerangProxy.Disable())

	_, err = c.Send(defaultMsg)
	require.ErrorIs(t, err, client.ErrNotRunning)

	_, _, err = c.SendSync(defaultCtx, defaultMsg)
	require.ErrorIs(t, err, client.ErrNotRunning)

	arbiter.RequireHappened("CLIENT_ONCLOSE_HOOK")
	arbiter.RequireHappened("CLIENT_WAS_CLOSING")
	arbiter.RequireHappened("CLIENT_WAS_CLOSED")
}
