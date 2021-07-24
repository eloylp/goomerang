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
	m1Name := message.FQDN(m1)
	r.Register(m1, func(serverOpts server.ServerOpts, msg proto.Message) error {
		return errors.New("this causes errors")
	})
	m2 := &testMessages.PingPong{Message: "ping"}
	m2Name := message.FQDN(m2)
	r.Register(m2, func(serverOpts server.ServerOpts, msg proto.Message) error {
		return nil
	})

	fakeServerOpts := &FakeServerOpts{}
	stubMsg := &testMessages.GreetV1{}

	msg, h1, err := r.Handler(m1Name)
	require.NoError(t, err)
	assert.Error(t, h1(fakeServerOpts, stubMsg))
	assert.Equal(t, m1Name, message.FQDN(msg))

	msg, h2, err := r.Handler(m2Name)
	require.NoError(t, err)
	assert.NoError(t, h2(fakeServerOpts, stubMsg))
	assert.Equal(t, m2Name, message.FQDN(msg))
}

type FakeServerOpts struct {
}

func (f *FakeServerOpts) Send(ctx context.Context, msg proto.Message) error {
	panic("implement me")
}

func (f *FakeServerOpts) Shutdown(ctx context.Context) error {
	panic("implement me")
}
