package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/test"
)

func TestFQDN(t *testing.T) {
	m := &test.GreetV1{}
	assert.Equal(t, "goomerang.test.GreetV1", message.FQDN(m))
}

func TestPackTimestamp(t *testing.T) {
	m := &test.GreetV1{}
	pack, err := message.Pack(m)
	require.NoError(t, err)
	unpack, err := message.UnPack(pack)
	require.NoError(t, err)
	now := time.Now().UnixMicro()
	packTime := unpack.Creation.AsTime().UnixMicro()
	assert.InDelta(t, now, packTime, 1000)
	assert.Less(t, packTime, now)
}
