package goomerang_test

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"go.eloylp.dev/kit/pki"
	"go.eloylp.dev/kit/test"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/server"
)

const (
	serverAddr = "0.0.0.0:3000"
)

func PrepareServer(t *testing.T, opts ...server.Option) (s *server.Server, run func()) {
	t.Helper()
	is := configureServer(t, opts)
	return is, func() {
		go is.Run()
		test.WaitTCPService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	}
}

func configureServer(t *testing.T, opts []server.Option) *server.Server {
	allOpts := []server.Option{server.WithListenAddr(serverAddr)}
	allOpts = append(allOpts, opts...)
	s, err := server.NewServer(allOpts...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func PrepareTLSServer(t *testing.T, opts ...server.Option) (s *server.Server, run func()) {
	t.Helper()
	is := configureServer(t, opts)
	return is, func() {
		go is.Run()
		test.WaitTLSService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	}
}

func PrepareClient(t *testing.T, opts ...client.Option) (c *client.Client, connect func()) {
	allOpts := []client.Option{client.WithTargetServer(serverAddr)}
	allOpts = append(allOpts, opts...)
	opts = append(allOpts)
	c, err := client.NewClient(opts...)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	return c, func() {
		err = c.Connect(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func SelfSignedCert(t *testing.T) tls.Certificate {
	crt, err := pki.SelfSignedCert(pki.WithCertSerialNumber(1),
		pki.WithCertCommonName("127.0.0.1"),
		pki.WithCertOrganization([]string{"goomerang"}),
		pki.WithCertIpAddresses([]string{"0.0.0.0"}),
		pki.WithCertNotBefore(time.Now()),
		pki.WithCertNotAfter(time.Now().Add(time.Hour*24*365)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return crt
}
