package message_test

import (
	"go.eloylp.dev/goomerang/internal/message"
	"go.eloylp.dev/goomerang/internal/message/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFQDN(t *testing.T) {
	m := &test.GreetV1{}
	assert.Equal(t, "goomerang.test.GreetV1", message.FQDN(m))
}