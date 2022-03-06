package messaging_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/messaging/test"
)

func TestMessageRegistry(t *testing.T) {
	msg := &test.GreetV1{}
	r := messaging.Registry{}
	r.Register("m1", msg)
	res, err := r.Message("m1")
	require.NoError(t, err)
	require.EqualValues(t, msg, res)
	require.NotSame(t, msg, res.(*test.GreetV1))
}

func TestMessageRegistryNotFound(t *testing.T) {
	r := messaging.Registry{}
	_, err := r.Message("m1")
	require.EqualError(t, err, "cannot found message with key: m1")
}
