//go:build integration

package goomerang_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestHandlerRegistrationMoment(t *testing.T) {
	t.Run("Client CAN register multiple handlers BEFORE run", func(t *testing.T) {
		c, err := client.New()
		require.NoError(t, err)
		registerClientDumbHandler(c)
		registerClientDumbHandler(c)
	})
	t.Run("Client CANNOT register handlers AFTER run", func(t *testing.T) {
		s, run := PrepareServer(t)
		run()
		defer s.Shutdown(defaultCtx)
		c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
		connect()
		defer c.Close(defaultCtx)
		assert.Panics(t, func() {
			registerClientDumbHandler(c)
		})
	})
	t.Run("Client CAN register multiple middlewares BEFORE run", func(t *testing.T) {
		c, err := client.New()
		require.NoError(t, err)
		registerClientDumbMiddleware(c)
		registerClientDumbMiddleware(c)
	})
	t.Run("Client CANNOT register middlewares AFTER run", func(t *testing.T) {
		s, run := PrepareServer(t)
		run()
		defer s.Shutdown(defaultCtx)
		c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
		connect()
		defer c.Close(defaultCtx)
		assert.Panics(t, func() {
			registerClientDumbMiddleware(c)
		})
	})
	t.Run("Server CAN register multiple handlers BEFORE run", func(t *testing.T) {
		s, err := server.New()
		require.NoError(t, err)
		registerServerDumbHandler(s)
		registerServerDumbHandler(s)
	})
	t.Run("Server CANNOT register handlers AFTER run", func(t *testing.T) {
		s, run := PrepareServer(t)
		run()
		defer s.Shutdown(defaultCtx)
		c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
		connect()
		defer c.Close(defaultCtx)
		assert.Panics(t, func() {
			registerServerDumbHandler(s)
		})
	})
	t.Run("Server CAN register multiple middlewares BEFORE run", func(t *testing.T) {
		s, err := server.New()
		require.NoError(t, err)
		registerServerDumbMiddleware(s)
		registerServerDumbMiddleware(s)
	})
	t.Run("Server CANNOT register middlewares AFTER run", func(t *testing.T) {
		s, run := PrepareServer(t)
		run()
		defer s.Shutdown(defaultCtx)
		c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
		connect()
		defer c.Close(defaultCtx)
		assert.Panics(t, func() {
			registerServerDumbMiddleware(s)
		})
	})
}

func registerClientDumbHandler(c *client.Client) {
	c.Handle(&protos.MessageV1{}, message.HandlerFunc(func(s message.Sender, w *message.Message) {}))
}

func registerClientDumbMiddleware(c *client.Client) {
	c.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, w *message.Message) {})
	})
}

func registerServerDumbHandler(s *server.Server) {
	s.Handle(&protos.MessageV1{}, message.HandlerFunc(func(s message.Sender, w *message.Message) {}))
}

func registerServerDumbMiddleware(s *server.Server) {
	s.Middleware(func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, w *message.Message) {})
	})
}
