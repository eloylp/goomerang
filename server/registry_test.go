package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
	testMessages "go.eloylp.dev/goomerang/message/test"
	"go.eloylp.dev/goomerang/server"
)

func TestHandlerRegistry(t *testing.T) {
	r := server.Registry{}

	m1 := &testMessages.GreetV1{Message: "hi!"}
	r.Register(m1, problematicHandler())
	m2 := &testMessages.PingPong{Message: "ping"}
	r.Register(m2, successfulHandler())

	fakeServerOpts := &FakeServerOpts{}

	m1Name := message.FQDN(m1)
	msg, h1, err := r.Handler(m1Name)
	require.NoError(t, err)
	assert.Error(t, h1[0](fakeServerOpts, m1))
	assert.Equal(t, m1Name, message.FQDN(msg))

	m2Name := message.FQDN(m2)
	msg, h2, err := r.Handler(m2Name)
	require.NoError(t, err)
	assert.NoError(t, h2[0](fakeServerOpts, m2))
	assert.Equal(t, m2Name, message.FQDN(msg))
}

func successfulHandler() server.ServerHandler {
	return func(serverOpts server.ServerOpts, msg proto.Message) error {
		return nil
	}
}

func problematicHandler() server.ServerHandler {
	return func(serverOpts server.ServerOpts, msg proto.Message) error {
		return errors.New("this causes errors")
	}
}

type FakeServerOpts struct {
}

func (f *FakeServerOpts) Send(ctx context.Context, msg proto.Message) error {
	panic("implement me")
}

func (f *FakeServerOpts) Shutdown(ctx context.Context) error {
	panic("implement me")
}

func TestMultipleCumulativeHandlersCanBeRegistered(t *testing.T) {
	r := server.Registry{}
	m := &testMessages.GreetV1{Message: "hi!"}
	r.Register(m, successfulHandler(), successfulHandler())
	r.Register(m, successfulHandler(), successfulHandler())
	_, handlers, err := r.Handler(message.FQDN(m))
	require.NoError(t, err)
	assert.Len(t, handlers, 4)
}
