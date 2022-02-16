package goomerang_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/message"
	testMessages "go.eloylp.dev/goomerang/internal/message/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

var defaultCtx = context.Background()

func TestPingPongServer(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	s.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.PingPong)
		if err := s.Send(defaultCtx, &message.Message{Payload: &testMessages.PingPong{Message: "pong"}}); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_PING")
	}))

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.PingPong)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_PONG")
	}))
	err := c.Send(defaultCtx, &message.Message{Payload: &testMessages.PingPong{Message: "ping"}})
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_PING")
	arbiter.RequireHappened("CLIENT_RECEIVED_PONG")
}

func TestSecuredPingPongServer(t *testing.T) {
	arbiter := test.NewArbiter(t)
	// Get self-signed certificate.
	certificate := SelfSignedCert(t)

	s := PrepareTLSServer(t, server.WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{certificate},
	}))

	defer s.Shutdown(defaultCtx)
	msg := &testMessages.PingPong{}

	s.RegisterHandler(msg, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.PingPong)
		if err := s.Send(defaultCtx, &message.Message{Payload: &testMessages.PingPong{Message: "pong"}}); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_PING")
	}))

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate.Leaf)
	c := PrepareClient(t, client.WithWithTLSConfig(&tls.Config{
		RootCAs: certPool,
	}))
	defer c.Close(defaultCtx)

	c.RegisterHandler(msg, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.PingPong)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_PONG")
	}))

	require.NoError(t, c.Send(defaultCtx, &message.Message{Payload: &testMessages.PingPong{Message: "ping"}}))
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_PING")
	arbiter.RequireHappened("CLIENT_RECEIVED_PONG")
}
