package goomerang_test

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/toxiproxy"
	toxiClient "github.com/Shopify/toxiproxy/client"
	"go.eloylp.dev/kit/pki"
	"go.eloylp.dev/kit/test"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/server"
)

const (
	proxyServerAddr   = "127.0.0.1:3000"
	proxyAddr         = "127.0.0.1:8474"
	serverBackendAddr = "127.0.0.1:3001"
)

var (
	proxyServer    *toxiproxy.ApiServer
	proxyClient    *toxiClient.Client
	goomerangProxy *toxiClient.Proxy
)

func init() {
	proxyServer = toxiproxy.NewServer()
	proxyAddrParts := strings.Split(proxyAddr, ":")

	go proxyServer.Listen(proxyAddrParts[0], proxyAddrParts[1])

	mustWaitTCPService("localhost:8474", time.Millisecond, time.Second)

	proxyClient = toxiClient.NewClient(proxyAddr)
	var err error
	goomerangProxy, err = proxyClient.CreateProxy("goomerang", proxyServerAddr, serverBackendAddr)
	if err != nil {
		panic(err)
	}
}

func mustWaitTCPService(addr string, interval, maxWait time.Duration) {
	ctx, cancl := context.WithTimeout(context.Background(), maxWait)
	defer cancl()
	for {
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		default:
			con, conErr := net.Dial("tcp", addr)
			if conErr == nil {
				_ = con.Close()
				return
			}
			time.Sleep(interval)
		}
	}
}

func PrepareServer(t *testing.T, opts ...server.Option) (s *server.Server, run func()) {
	t.Helper()
	is := configureServer(t, opts)
	return is, func() {
		go is.Run()
		test.WaitTCPService(t, serverBackendAddr, 50*time.Millisecond, 2*time.Second)
	}
}

func configureServer(t *testing.T, opts []server.Option) *server.Server {
	allOpts := []server.Option{server.WithListenAddr(serverBackendAddr)}
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
		test.WaitTLSService(t, serverBackendAddr, 50*time.Millisecond, 2*time.Second)
	}
}

func PrepareClient(t *testing.T, opts ...client.Option) (c *client.Client, connect func()) {
	allOpts := []client.Option{client.WithTargetServer(proxyServerAddr)}
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
		pki.WithCertIpAddresses([]string{"127.0.0.1"}),
		pki.WithCertNotBefore(time.Now()),
		pki.WithCertNotAfter(time.Now().Add(time.Hour*24*365)),
	)
	if err != nil {
		t.Fatal(err)
	}
	return crt
}
