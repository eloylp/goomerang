package rpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client/internal/rpc"
)

func TestRPCRegistry(t *testing.T) {
	reg := rpc.NewRegistry()
	id := "09AF"

	reg.CreateListener(id)
	m := &rpc.MultiReply{}

	err := reg.SubmitResult(id, m)
	require.NoError(t, err)

	result, err := reg.ResultFor(context.Background(), id)
	require.NoError(t, err)

	assert.Same(t, m, result)

	err = reg.SubmitResult(id, m)
	assert.Errorf(t, err, "rpc-registry: cannot find key for 09AF", "Last r.ResultFor key should remove key entry")
}

func TestRPCRegistry_SubmitResult(t *testing.T) {
	reg := rpc.NewRegistry()
	err := reg.SubmitResult("NON_EXISTENT", &rpc.MultiReply{})
	assert.Errorf(t, err, "rpc-registry: cannot find key for NON_EXISTENT")
}

func TestRPCRegistry_ResultFor(t *testing.T) {
	reg := rpc.NewRegistry()
	_, err := reg.ResultFor(context.Background(), "NON_EXISTENT")
	assert.Errorf(t, err, "rpc-registry: cannot find result for key for NON_EXISTENT")
}

func TestRegistry_ResultFor_WaitsUntilResultArrives(t *testing.T) {
	reg := rpc.NewRegistry()
	reg.CreateListener("09AF")

	reply := &rpc.MultiReply{}

	time.AfterFunc(time.Millisecond*500, func() {
		_ = reg.SubmitResult("09AF", reply)
	})
	result, err := reg.ResultFor(context.Background(), "09AF")
	require.NoError(t, err)
	assert.Same(t, reply, result)
}

func TestRegistry_ResultFor_cancelOncontext(t *testing.T) {
	reg := rpc.NewRegistry()
	reg.CreateListener("09AF")

	reply := &rpc.MultiReply{}

	time.AfterFunc(time.Millisecond*500, func() {
		_ = reg.SubmitResult("09AF", reply)
	})
	ctx, cancl := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancl()

	_, err := reg.ResultFor(ctx, "09AF")
	require.EqualError(t, err, "rpcregistry: context deadline exceeded")
}
