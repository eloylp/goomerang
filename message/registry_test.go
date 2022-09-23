//go:build unit

package message_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/message"
)

func TestMessageRegistry(t *testing.T) {
	msg := &protos.MessageV1{}
	r := message.Registry{}

	r.Register("m1", msg)
	res, err := r.Message("m1")

	require.NoError(t, err)
	require.EqualValues(t, msg, res)
	require.NotSame(t, msg, res.(*protos.MessageV1))
}

func TestMessageRegistryNotFound(t *testing.T) {
	r := message.Registry{}
	_, err := r.Message("m1")
	require.EqualError(t, err, "cannot found message with key: m1")
}
