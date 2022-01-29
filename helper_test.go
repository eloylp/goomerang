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

func PrepareServer(t *testing.T, opts ...server.Option) *server.Server {
	t.Helper()
	s := configureServer(t, opts)
	go s.Run()
	test.WaitTCPService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	return s
}

func configureServer(t *testing.T, opts []server.Option) *server.Server {
	opts = append(opts, server.WithListenAddr(serverAddr))
	s, err := server.NewServer(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func PrepareTLSServer(t *testing.T, opts ...server.Option) *server.Server {
	t.Helper()
	s := configureServer(t, opts)
	go s.Run()
	test.WaitTLSService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
	return s
}

func PrepareClient(t *testing.T, opts ...client.Option) *client.Client {
	opts = append(opts, client.WithTargetServer(serverAddr))
	c, err := client.NewClient(opts...)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = c.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return c
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
