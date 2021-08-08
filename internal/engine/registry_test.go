package engine_test

import (
	"errors"
	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/engine"
)

type fakeHandler func() error

func TestHandlerRegistry(t *testing.T) {
	r := engine.Registry{}

	m1 := &testMessages.GreetV1{Message: "hi!"}
	r.Register(m1, problematicHandler())
	m2 := &testMessages.PingPong{Message: "ping"}
	r.Register(m2, successfulHandler())

	m1Name := message.FQDN(m1)
	msg, h1, err := r.Handler(m1Name)
	require.NoError(t, err)
	assert.Error(t, h1[0].(fakeHandler)())
	assert.Equal(t, m1Name, message.FQDN(msg))

	m2Name := message.FQDN(m2)
	msg, h2, err := r.Handler(m2Name)
	require.NoError(t, err)
	assert.NoError(t, h2[0].(fakeHandler)())
	assert.Equal(t, m2Name, message.FQDN(msg))
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
	r := engine.Registry{}
	m := &testMessages.GreetV1{Message: "hi!"}
	r.Register(m, successfulHandler(), successfulHandler())
	r.Register(m, successfulHandler(), successfulHandler())
	_, handlers, err := r.Handler(message.FQDN(m))
	require.NoError(t, err)
	assert.Len(t, handlers, 4)
}
