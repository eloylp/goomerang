package server

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Cfg)

// WithListenAddr the address in which the server
// will listen to client connections.
func WithListenAddr(addr string) Option {
	return func(cfg *Cfg) {
		cfg.ListenURL = addr
	}
}

// WithOnStatusChangeHook allows the user registering function
// hooks for each status changes. Possible values for statuses
// can be found at go.eloylp.dev/goomerang/ws package.
//
// Sending multiple times this option in the constructor will lead
// to function concatenation, so multiple hooks can be configured.
//
// IMPORTANT NOTE: There's no panic catcher implementation here,
// ensure to implement it or ensure the underlying code does not
// panic.
func WithOnStatusChangeHook(h func(status uint32)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnStatusChange(h)
	}
}

// WithOnCloseHook allows the user registering function
// hooks for when the server reaches the ws.StatusClosed status.
//
// Sending multiple times this option in the constructor will lead
// to function concatenation, so multiple hooks can be configured.
//
// IMPORTANT NOTE: There's no panic catcher implementation here,
// ensure to implement it or ensure the underlying code does not
// panic.
func WithOnCloseHook(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnClose(h)
	}
}

// WithOnErrorHook allows the user registering function
// hooks for when an internal error happens, but cannot be
// returned in any way to the user. Like in processing loops.
//
// IMPORTANT NOTE: There's no panic catcher implementation here,
// ensure to implement it or ensure the underlying code does not
// panic.
func WithOnErrorHook(h func(err error)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnError(h)
	}
}

// WithOnConfiguration allows the user registering hook
// functions which will be invoked once the server is configured.
func WithOnConfiguration(h func(cfg *Cfg)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnConfiguration(h)
	}
}

// WithOnWorkerStart allows the user registering hook
// functions which will be invoked whenever a new
// handler goroutine is invoked. Only if concurrency
// is enabled by the WithMaxConcurrency option.
func WithOnWorkerStart(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnWorkerStart(h)
	}
}

// WithOnWorkerEnd allows the user registering hook
// functions which will be invoked whenever a
// handler goroutine ends its operations. Only if concurrency
// is enabled by using WithMaxConcurrency option.
func WithOnWorkerEnd(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnWorkerEnd(h)
	}
}

// WithOnBroadcastHook allows the user to inject a hook which
// will be executed on each successfully broadcast operation.
func WithOnBroadcastHook(f func(fqdn string, result []BroadcastResult, duration time.Duration)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnBroadcast(f)
	}
}

// WithOnClientBroadcastHook allows the user to inject a hook which
// will be executed each successfully broadcast triggered
// by a client broadcast command.
func WithOnClientBroadcastHook(f func(fqdn string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnClientBroadcast(f)
	}
}

// WithOnSubscribeHook allows the user to inject a hook which
// will be executed each successfully subscribe
// to a specific topic, for a message.
func WithOnSubscribeHook(f func(topic string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnSubscribe(f)
	}
}

// WithOnUnsubscribeHook allows the user to inject a hook which
// will be executed each time successfully unsubscribe
// from a specific topic, for a message.
func WithOnUnsubscribeHook(f func(topic string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnUnsubscribe(f)
	}
}

// WithOnPublishHook allows the user to inject a hook which
// will be executed each time successfully publish operation
// takes place for a specific topic and message (fqdn).
func WithOnPublishHook(f func(topic, fqdn string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnPublish(f)
	}
}

// WithMaxConcurrency sets the concurrency level for handler
// execution. Values <= 1 means no concurrency, which means
// no goroutine scheduling takes place. Defaults to 10.
func WithMaxConcurrency(n int) Option {
	return func(cfg *Cfg) {
		cfg.MaxConcurrency = n
	}
}

// WithTLSConfig allows the user to pass a *tls.Config
// full setup to the server, in which encryption and authentication
// could be configured, by setting up a PKI (public key infrastructure).
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *Cfg) {
		cfg.TLSConfig = tlsConfig
	}
}

// WithReadBufferSize configures the size in bytes of the
// read buffers for each connection. Defaults on 4096 bytes.
func WithReadBufferSize(s int) Option {
	return func(cfg *Cfg) {
		cfg.ReadBufferSize = s
	}
}

// WithWriteBufferSize configures the size in bytes of the
// write buffers for each connection. Defaults on 4096 bytes.
func WithWriteBufferSize(s int) Option {
	return func(cfg *Cfg) {
		cfg.WriteBufferSize = s
	}
}

// WithCompressionEnabled specifies if the server should try
// to negotiate compression (see https://datatracker.ietf.org/doc/html/rfc7692).
// Enabling this does not guarantee its success.
func WithCompressionEnabled(b bool) Option {
	return func(cfg *Cfg) {
		cfg.EnableCompression = b
	}
}

// WithHandShakeTimeout set the time limit in which the websocket
// handshake should take place.
func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HandshakeTimeout = d
	}
}

// WithHTTPWriteTimeout corresponds to the
// http.Server.WriteTimeout configuration for
// this websocket server.
func WithHTTPWriteTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPWriteTimeout = t
	}
}

// WithHTTPReadTimeout correspond to the
// http.Server.ReadTimeout configuration for
// this websocket server.
func WithHTTPReadTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPReadTimeout = t
	}
}

// WithHTTPReadHeaderTimeout corresponds to the
// http.Server.ReadHeaderTimeout for
// this websocket server.
func WithHTTPReadHeaderTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPReadHeaderTimeout = t
	}
}
