package message_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/message/test"
)

func TestFQDN(t *testing.T) {
	m := &test.GreetV1{}
	assert.Equal(t, "goomerang.test.GreetV1", message.FQDN(m))
}
