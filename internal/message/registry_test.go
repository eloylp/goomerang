package message_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/message"
)

type fakeHandler func() error

func TestHandlerRegistry(t *testing.T) {
	r := message.Registry{}

	r.Register("m1", problematicHandler())
	r.Register("m2", successfulHandler())

	h1, err := r.Handler("m1")
	require.NoError(t, err)
	assert.Error(t, h1[0].(fakeHandler)())

	h2, err := r.Handler("m2")
	require.NoError(t, err)
	assert.NoError(t, h2[0].(fakeHandler)())
}

func successfulHandler() fakeHandler {
	return func() error {
		return nil
	}
}

func problematicHandler() fakeHandler {
	return func() error {
		return errors.New("this causes errors")
	}
}

func TestMultipleCumulativeHandlersCanBeRegistered(t *testing.T) {
	r := message.Registry{}
	r.Register("m", successfulHandler(), successfulHandler())
	r.Register("m", successfulHandler(), successfulHandler())
	handlers, err := r.Handler("m")
	require.NoError(t, err)
	assert.Len(t, handlers, 4)
}
