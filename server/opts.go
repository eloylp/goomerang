package server

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Cfg)

func WithListenAddr(addr string) Option {
	return func(cfg *Cfg) {
		cfg.ListenURL = addr
	}
}

func WithOnStatusChangeHook(h func(status uint32)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnStatusChange(h)
	}
}

func WithOnCloseHook(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnClose(h)
	}
}

func WithOnErrorHook(h func(err error)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnError(h)
	}
}

func WithOnConfiguration(h func(cfg *Cfg)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnConfiguration(h)
	}
}

func WithOnWorkerStart(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnWorkerStart(h)
	}
}

func WithOnWorkerEnd(h func()) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnWorkerEnd(h)
	}
}

func WithMaxConcurrency(n int) Option {
	return func(cfg *Cfg) {
		cfg.MaxConcurrency = n
	}
}

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *Cfg) {
		cfg.TLSConfig = tlsConfig
	}
}

func WithReadBufferSize(s int) Option {
	return func(cfg *Cfg) {
		cfg.ReadBufferSize = s
	}
}

func WithWriteBufferSize(s int) Option {
	return func(cfg *Cfg) {
		cfg.WriteBufferSize = s
	}
}

func WithHTTPWriteTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPWriteTimeout = t
	}
}

func WithHTTPReadTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPReadTimeout = t
	}
}

func WithHTTPReadHeaderTimeout(t time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HTTPReadHeaderTimeout = t
	}
}

func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HandshakeTimeout = d
	}
}

func WithCompressionEnabled(b bool) Option {
	return func(cfg *Cfg) {
		cfg.EnableCompression = b
	}
}
