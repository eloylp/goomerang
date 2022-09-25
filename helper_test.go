package goomerang_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/toxiproxy"
	toxiClient "github.com/Shopify/toxiproxy/client"
	"go.eloylp.dev/kit/pki"
	kitTest "go.eloylp.dev/kit/test"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	message "go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
	"go.eloylp.dev/goomerang/ws"
)

const (
	proxyAddr         = "127.0.0.1:8474"
	kernelDefinedPort = "127.0.0.1:0"
)

var (
	proxyServer *toxiproxy.ApiServer
	proxyClient *toxiClient.Client
	defaultCtx  = context.Background()
	defaultMsg  = func() *message.Message {
		return message.New().
			SetPayload(&protos.MessageV1{
				Message: "a message!",
			})
	}
	echoHandler = message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		s.Send(msg)
	})
	nilHandler = message.HandlerFunc(func(s message.Sender, msg *message.Message) {})
)

func init() {
	proxyServer = toxiproxy.NewServer()
	proxyAddrParts := strings.Split(proxyAddr, ":")
	go proxyServer.Listen(proxyAddrParts[0], proxyAddrParts[1])
	mustWaitTCPService("localhost:8474", time.Millisecond, time.Second)
	proxyClient = toxiClient.NewClient(proxyAddr)
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

func Server(t *testing.T, opts ...server.Option) (s *server.Server, run func()) {
	t.Helper()
	is := configureServer(t, opts)
	return is, func() {
		go is.Run()
		waitForServer(t, is)
	}
}

func waitForServer(t *testing.T, is *server.Server) {
	ctx, cancl := context.WithTimeout(defaultCtx, time.Second)
	defer cancl()
	var serverAddr string
	for {
		select {
		case <-ctx.Done():
			t.Fatal(fmt.Errorf("error waiting for server addr: %v", ctx.Err()))
		default:
			serverAddr = is.Addr()
			if serverAddr == "" {
				continue
			}
			kitTest.WaitTCPService(t, serverAddr, 50*time.Millisecond, 2*time.Second)
			return
		}
	}
}

func configureServer(t *testing.T, opts []server.Option) *server.Server {
	allOpts := []server.Option{server.WithListenAddr(kernelDefinedPort)}
	allOpts = append(allOpts, opts...)
	s, err := server.New(allOpts...)
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
		waitForServer(t, is)
	}
}

func Client(t *testing.T, opts ...client.Option) (c *client.Client, connect func()) {
	c, err := client.New(opts...)
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

func statusChangesHook(a *test.Arbiter, side string) func(status uint32) {
	side = strings.ToUpper(side)
	return func(status uint32) {
		switch status {
		case ws.StatusNew:
			a.ItsAFactThat(side + "_WAS_NEW")
		case ws.StatusRunning:
			a.ItsAFactThat(side + "_WAS_RUNNING")
		case ws.StatusClosing:
			a.ItsAFactThat(side + "_WAS_CLOSING")
		case ws.StatusClosed:
			a.ItsAFactThat(side + "_WAS_CLOSED")
		}
	}
}

func noErrorHook(a *test.Arbiter) func(err error) {
	return func(err error) {
		a.ErrorHappened(err)
	}
}

func failIfErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
