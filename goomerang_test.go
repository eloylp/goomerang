package goomerang_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	testMessages "go.eloylp.dev/goomerang/internal/messaging/test"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

var defaultCtx = context.Background()

func TestRoundTrip(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	s.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.MessageV1)
		if _, err := s.Send(&message.Message{Payload: &testMessages.MessageV1{Message: "pong"}}); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_MSG")
	}))
	run()
	c, connect := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	payloadSize, err := c.Send(&message.Message{Payload: &testMessages.MessageV1{Message: "ping"}})
	require.NoError(t, err)
	require.NotEmpty(t, payloadSize)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG")
	arbiter.RequireHappened("CLIENT_RECEIVED_REPLY")
}

func TestSecuredRoundTrip(t *testing.T) {
	arbiter := test.NewArbiter(t)
	// Get self-signed certificate.
	certificate := SelfSignedCert(t)

	s, run := PrepareTLSServer(t, server.WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{certificate},
	}))

	defer s.Shutdown(defaultCtx)
	msg := &testMessages.MessageV1{}

	s.RegisterHandler(msg, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.MessageV1)
		if _, err := s.Send(&message.Message{Payload: &testMessages.MessageV1{Message: "pong"}}); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_MSG")
	}))
	run()
	certPool := x509.NewCertPool()
	certPool.AddCert(certificate.Leaf)
	c, connect := PrepareClient(t, client.WithWithTLSConfig(&tls.Config{
		RootCAs: certPool,
	}))
	defer c.Close(defaultCtx)

	c.RegisterHandler(msg, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*testMessages.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	payloadSize, err := c.Send(&message.Message{Payload: &testMessages.MessageV1{Message: "ping"}})
	assert.NoError(t, err)
	assert.NotEmpty(t, payloadSize)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG")
	arbiter.RequireHappened("CLIENT_RECEIVED_REPLY")
}

func TestMiddlewares(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)
	s.RegisterMiddleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("SERVER_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})
	s.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("SERVER_HANDLER_EXECUTED")
		if _, err := s.Send(&message.Message{
			Payload: msg.Payload,
		}); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()
	c, connect := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterMiddleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("CLIENT_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})

	c.RegisterHandler(&testMessages.MessageV1{}, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	_, err := c.Send(&message.Message{Payload: &testMessages.MessageV1{Message: "ping"}})
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappenedInOrder("SERVER_MIDDLEWARE_EXECUTED", "SERVER_HANDLER_EXECUTED")
	arbiter.RequireHappenedInOrder("SERVER_HANDLER_EXECUTED", "CLIENT_MIDDLEWARE_EXECUTED")
	arbiter.RequireHappenedInOrder("CLIENT_MIDDLEWARE_EXECUTED", "CLIENT_RECEIVED_REPLY")
}

func TestHeadersAreSent(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t)
	defer s.Shutdown(defaultCtx)

	m := &testMessages.MessageV1{}
	s.RegisterHandler(m, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" { //nolint: goconst
			arbiter.ItsAFactThat("SERVER_RECEIVED_MSG_HEADERS")
		}
		if _, err := s.Send(msg); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()

	c, connect := PrepareClient(t)
	defer c.Close(defaultCtx)

	c.RegisterHandler(m, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" {
			arbiter.ItsAFactThat("CLIENT_RECEIVED_MSG_HEADERS")
		}
	}))
	connect()
	msg := &message.Message{
		Payload: &testMessages.MessageV1{Message: "ping"},
		Header:  map[string]string{"h1": "v1"},
	}
	_, err := c.Send(msg)
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG_HEADERS")
	arbiter.RequireHappened("CLIENT_RECEIVED_MSG_HEADERS")
}
