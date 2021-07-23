package goomerang_test

import (
	"errors"
	"go.eloylp.dev/goomerang/message"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang"
)

func TestHandlerRegistry(t *testing.T) {
	r := goomerang.Registry{}

	m1Name := "m1"
	r.Register(m1Name, func(serverOpts goomerang.ServerOpts, msg proto.Message) error {
		return errors.New("this causes errors")
	})
	m2Name := "m2"
	r.Register(m2Name, func(serverOpts goomerang.ServerOpts, msg proto.Message) error {
		return nil
	})

	fakeServerOpts := &FakeServerOpts{}
	stubMsg := &message.GreetV1{}

	h1, err := r.Handler(m1Name)
	require.NoError(t, err)
	assert.Error(t, h1(fakeServerOpts, stubMsg))

	h2, err := r.Handler(m2Name)
	require.NoError(t, err)
	assert.NoError(t, h2(fakeServerOpts, stubMsg))

}
