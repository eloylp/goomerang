package message_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/test"
)

func TestMessageRegistry(t *testing.T) {
	msg := &test.GreetV1{Message: "hi!"}
	r := message.MessageRegistry{}
	r.Register("m1", msg)
	res, err := r.Message("m1")
	require.NoError(t, err)
	require.Equal(t, msg, res)
}

func TestMessageRegistryNotFound(t *testing.T) {
	r := message.MessageRegistry{}
	_, err := r.Message("m1")
	require.EqualError(t, err, "cannot found message with key: m1")
}
