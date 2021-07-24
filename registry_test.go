package goomerang_test

import (
	"errors"
	message "go.eloylp.dev/goomerang/message/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang"
)

func TestHandlerRegistry(t *testing.T) {
	r := goomerang.Registry{}

	m1 := &message.GreetV1{Message: "hi!"}
	m1Name := m1.ProtoReflect().Descriptor().FullName()
	r.Register(m1, func(serverOpts goomerang.ServerOpts, msg proto.Message) error {
		return errors.New("this causes errors")
	})
	m2 := &message.PingPong{Message: "ping"}
	m2Name := m2.ProtoReflect().Descriptor().FullName()
	r.Register(m2, func(serverOpts goomerang.ServerOpts, msg proto.Message) error {
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
