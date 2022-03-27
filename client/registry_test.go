package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/message"
)

func TestRequestRegistry(t *testing.T) {
	reg := newRegistry()
	id := "09AF"

	reg.createListener(id)
	m := &message.Message{}

	err := reg.submitResult(id, m)
	require.NoError(t, err)

	result, err := reg.resultFor(context.Background(), id)
	require.NoError(t, err)

	assert.Same(t, m, result)

	err = reg.submitResult(id, m)
	assert.Errorf(t, err, "request registry: cannot find key for 09AF", "Last r.ResultFor key should remove key entry")
}

func TestRequestRegistry_SubmitResult(t *testing.T) {
	reg := newRegistry()
	err := reg.submitResult("NON_EXISTENT", &message.Message{})
	assert.Errorf(t, err, "request registry: cannot find key for NON_EXISTENT")
}

func TestRequestRegistry_ResultFor(t *testing.T) {
	reg := newRegistry()
	_, err := reg.resultFor(context.Background(), "NON_EXISTENT")
	assert.Errorf(t, err, "request registry: cannot find result for key for NON_EXISTENT")
}

func TestRequestRegistry_ResultFor_WaitsUntilResultArrives(t *testing.T) {
	reg := newRegistry()
	reg.createListener("09AF")

	reply := &message.Message{}

	time.AfterFunc(time.Millisecond*500, func() {
		_ = reg.submitResult("09AF", reply)
	})
	result, err := reg.resultFor(context.Background(), "09AF")
	require.NoError(t, err)
	assert.Same(t, reply, result)
}

func TestRequestRegistry_ResultFor_cancelOncontext(t *testing.T) {
	reg := newRegistry()
	reg.createListener("09AF")

	reply := &message.Message{}

	time.AfterFunc(time.Millisecond*500, func() {
		_ = reg.submitResult("09AF", reply)
	})
	ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancl()

	_, err := reg.resultFor(ctx, "09AF")
	require.EqualError(t, err, "request registry: context deadline exceeded")
}
