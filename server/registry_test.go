package server_test

import (
	"context"
	"errors"
	message "go.eloylp.dev/goomerang/message/test"
	"go.eloylp.dev/goomerang/server"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandlerRegistry(t *testing.T) {
	r := server.Registry{}

	m1 := &message.GreetV1{Message: "hi!"}
	m1Name := m1.ProtoReflect().Descriptor().FullName()
	r.Register(m1, func(serverOpts server.ServerOpts, msg proto.Message) error {
		return errors.New("this causes errors")
	})
	m2 := &message.PingPong{Message: "ping"}
	m2Name := m2.ProtoReflect().Descriptor().FullName()
	r.Register(m2, func(serverOpts server.ServerOpts, msg proto.Message) error {
		return nil
	})

	fakeServerOpts := &FakeServerOpts{}
	stubMsg := &message.GreetV1{}

	msg, h1, err := r.Handler(string(m1Name))
	require.NoError(t, err)
	assert.Error(t, h1(fakeServerOpts, stubMsg))
	assert.Equal(t, m1Name, msg.ProtoReflect().Descriptor().FullName())

	msg, h2, err := r.Handler(string(m2Name))
	require.NoError(t, err)
	assert.NoError(t, h2(fakeServerOpts, stubMsg))
	assert.Equal(t, m2Name, msg.ProtoReflect().Descriptor().FullName())
}

type FakeServerOpts struct {
}

func (f *FakeServerOpts) Send(ctx context.Context, msg proto.Message) error {
	panic("implement me")
}

func (f *FakeServerOpts) Shutdown(ctx context.Context) error {
	panic("implement me")
}