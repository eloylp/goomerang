package server

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *cfg)

func WithListenAddr(addr string) Option {
	return func(cfg *cfg) {
		cfg.listenURL = addr
	}
}

func WithOnStatusChangeHook(h func(status uint32)) Option {
	return func(cfg *cfg) {
		cfg.hooks.AppendOnStatusChange(h)
	}
}

func WithOnCloseHook(h func()) Option {
	return func(cfg *cfg) {
		cfg.hooks.AppendOnClose(h)
	}
}

func WithOnErrorHook(h func(err error)) Option {
	return func(cfg *cfg) {
		cfg.hooks.AppendOnError(h)
	}
}

func WithMaxConcurrency(n int) Option {
	return func(cfg *cfg) {
		cfg.maxConcurrency = n
	}
}

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *cfg) {
		cfg.tlsConfig = tlsConfig
	}
}

func WithReadBufferSize(s int) Option {
	return func(cfg *cfg) {
		cfg.readBufferSize = s
	}
}

func WithWriteBufferSize(s int) Option {
	return func(cfg *cfg) {
		cfg.writeBufferSize = s
	}
}

func WithHTTPWriteTimeout(t time.Duration) Option {
	return func(cfg *cfg) {
		cfg.httpWriteTimeout = t
	}
}

func WithHTTPReadTimeout(t time.Duration) Option {
	return func(cfg *cfg) {
		cfg.httpReadTimeout = t
	}
}

func WithHTTPReadHeaderTimeout(t time.Duration) Option {
	return func(cfg *cfg) {
		cfg.httpReadHeaderTimeout = t
	}
}

func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *cfg) {
		cfg.handshakeTimeout = d
	}
}

func WithCompressionEnabled(b bool) Option {
	return func(cfg *cfg) {
		cfg.enableCompression = b
	}
}
