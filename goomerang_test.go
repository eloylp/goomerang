//go:build integration

package goomerang_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestRoundTrip(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := Server(t)
	defer s.Shutdown(defaultCtx)

	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		if _, err := s.Send(defaultMsg); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_MSG")
	}))
	run()
	c, connect := Client(t, client.WithServerAddr(s.Addr()))
	defer c.Close(defaultCtx)

	c.Handle(defaultMsg.Payload, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	payloadSize, err := c.Send(defaultMsg)
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
	msg := defaultMsg.Payload

	s.Handle(msg, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		if _, err := s.Send(defaultMsg); err != nil {
			arbiter.ErrorHappened(err)
		}
		arbiter.ItsAFactThat("SERVER_RECEIVED_MSG")
	}))
	run()
	certPool := x509.NewCertPool()
	certPool.AddCert(certificate.Leaf)
	c, connect := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithTLSConfig(&tls.Config{
			RootCAs: certPool,
		}))
	defer c.Close(defaultCtx)

	c.Handle(msg, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	payloadSize, err := c.Send(defaultMsg)
	assert.NoError(t, err)
	assert.NotEmpty(t, payloadSize)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG")
	arbiter.RequireHappened("CLIENT_RECEIVED_REPLY")
}

func TestMiddlewares(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := Server(t)
	defer s.Shutdown(defaultCtx)
	s.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("SERVER_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})
	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("SERVER_HANDLER_EXECUTED")
		if _, err := s.Send(&message.Message{
			Payload: msg.Payload,
		}); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()
	c, connect := Client(t, client.WithServerAddr(s.Addr()))
	defer c.Close(defaultCtx)

	c.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			arbiter.ItsAFactThat("CLIENT_MIDDLEWARE_EXECUTED")
			h.Handle(s, msg)
		})
	})

	c.Handle(defaultMsg.Payload, message.HandlerFunc(func(c message.Sender, msg *message.Message) {
		arbiter.ItsAFactThat("CLIENT_RECEIVED_REPLY")
	}))
	connect()
	_, err := c.Send(defaultMsg)
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappenedInOrder(
		"SERVER_MIDDLEWARE_EXECUTED",
		"SERVER_HANDLER_EXECUTED",
		"CLIENT_MIDDLEWARE_EXECUTED",
		"CLIENT_RECEIVED_REPLY",
	)
}

func TestHeadersAreSent(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := Server(t)
	defer s.Shutdown(defaultCtx)

	m := defaultMsg.Payload
	s.Handle(m, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" { //nolint: goconst
			arbiter.ItsAFactThat("SERVER_RECEIVED_MSG_HEADERS")
		}
		if _, err := s.Send(msg); err != nil {
			arbiter.ErrorHappened(err)
		}
	}))
	run()

	c, connect := Client(t, client.WithServerAddr(s.Addr()))
	defer c.Close(defaultCtx)

	c.Handle(m, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		if msg.Header.Get("h1") == "v1" {
			arbiter.ItsAFactThat("CLIENT_RECEIVED_MSG_HEADERS")
		}
	}))
	connect()
	msg := message.New().
		SetPayload(&protos.MessageV1{Message: "ping"}).
		SetHeader("h1", "v1")

	_, err := c.Send(msg)
	require.NoError(t, err)
	arbiter.RequireNoErrors()
	arbiter.RequireHappened("SERVER_RECEIVED_MSG_HEADERS")
	arbiter.RequireHappened("CLIENT_RECEIVED_MSG_HEADERS")
}

func TestUserCanAccessServerRegistry(t *testing.T) {
	s, _ := Server(t)
	s.Handle(defaultMsg.Payload, nilHandler)
	msg, err := s.Registry().Message(defaultMsg.Metadata.Kind)
	require.NoError(t, err)
	assert.Equal(t, messaging.FQDN(defaultMsg.Payload), messaging.FQDN(msg))
}

func TestUserCanConfigureCustomHTTPRoutes(t *testing.T) {
	t.Parallel()

	s, err := server.New(server.WithListenAddr("127.0.0.1:0"))
	require.NoError(t, err)

	s.Router().Handle("/my-endpoint", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hi!"))
	}))

	// Run the server
	go func() {
		_ = s.Run()
	}()
	waitForServer(t, s)
	defer s.Shutdown(defaultCtx)

	// Make the request to the custom endpoint
	url := fmt.Sprintf("http://%s/my-endpoint", s.Addr())
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "hi!", string(body))
}

func TestUserCannotConfigureCustomHTTPRoutesOnceRunning(t *testing.T) {
	t.Parallel()

	s, run := Server(t)
	run()
	defer s.Shutdown(defaultCtx)
	assert.Panics(t, func() {
		s.Router()
	})
}
