package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/security"
)

func TestSelfSignedPair(t *testing.T) {
	certificate, err := security.SelfSignedCertificate()
	require.NoError(t, err)
	assert.NotEmpty(t, certificate.Certificate)
	assert.NotEmpty(t, certificate.PrivateKey)
	assert.NotEmpty(t, certificate.Leaf)
}
