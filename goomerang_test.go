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

func TestMiddlewares(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	s.RegisterMiddleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("SERVER_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})
	s.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("SERVER_HANDLER_EXECUTED")
		if err := s.Send(context.Background(), &message.Message{
			Payload: msg.Payload,
		}); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterMiddleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("CLIENT_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})

	c.RegisterHandler(&testMessages.PingPong{}, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT_RECEIVED_PONG")
	}))
	err := c.Send(defaultCtx, &message.Message{Payload: &testMessages.PingPong{Message: "ping"}})
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappenedInOrder("SERVER_MIDDLEWARE_EXECUTED", "SERVER_HANDLER_EXECUTED")
	arbiter.RequireHappenedInOrder("SERVER_HANDLER_EXECUTED", "CLIENT_MIDDLEWARE_EXECUTED")
	arbiter.RequireHappenedInOrder("CLIENT_MIDDLEWARE_EXECUTED", "CLIENT_RECEIVED_PONG")
}

func TestHeadersAreSent(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	m := &testMessages.PingPong{}
	s.RegisterHandler(m, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" { //nolint: goconst
			arbiter.ItsAFactThat("SERVER_RECEIVED_MSG_HEADERS")
		}
		if err := s.Send(defaultCtx, msg); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))

	c := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterHandler(m, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" {
			arbiter.ItsAFactThat("CLIENT_RECEIVED_MSG_HEADERS")
		}
	}))
	msg := &message.Message{
		Payload: &testMessages.PingPong{Message: "ping"},
		Header:  map[string]string{"h1": "v1"},
	}
	err := c.Send(defaultCtx, msg)
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG_HEADERS")
	arbiter.RequireHappened("CLIENT_RECEIVED_MSG_HEADERS")
}
