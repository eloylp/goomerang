package client

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Cfg)

// WithServerAddr configures the server to
// connect to.
func WithServerAddr(addr string) Option {
	return func(cfg *Cfg) {
		cfg.TargetServer = addr
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

// WithHeartbeatInterval introduces how often the Ping/Pong
// operation should take place. Right now, this is intended
// to keep alive connections. Defaults to 5 seconds.
func WithHeartbeatInterval(interval time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HeartbeatInterval = interval
	}
}

// WithOnCloseHook allows the user registering function
// hooks for when the client reaches the ws.StatusClosed status.
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
// returned in any way to the client. Like in processing loops.
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
// functions which will be invoked once the client is configured.
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

// WithMaxConcurrency sets the concurrency level for handler
// execution. Values <= 1 means no concurrency, which means
// no goroutine scheduling takes place. Defaults to 10.
func WithMaxConcurrency(n int) Option {
	return func(cfg *Cfg) {
		cfg.MaxConcurrency = n
	}
}

// WithTLSConfig allows the user to pass a *tls.Config
// full setup to the client, in which encryption and authentication
// could be configured, by setting up a PKI (public key infrastructure).
func WithTLSConfig(tlsCfg *tls.Config) Option {
	return func(cfg *Cfg) {
		cfg.TLSConfig = tlsCfg
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

// WithCompressionEnabled specifies if the client should try
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
