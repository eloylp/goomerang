package client

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *cfg)

func WithTargetServer(addr string) Option {
	return func(cfg *cfg) {
		cfg.targetServer = addr
	}
}

func WithOnStatusChangeHook(h func(status uint32)) Option {
	return func(cfg *cfg) {
		cfg.hooks.AppendOnStatusChange(h)
	}
}

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(cfg *cfg) {
		cfg.heartbeatInterval = interval
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

func WithWithTLSConfig(tlsCfg *tls.Config) Option {
	return func(cfg *cfg) {
		cfg.tlsConfig = tlsCfg
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

func WithCompressionEnabled(b bool) Option {
	return func(cfg *cfg) {
		cfg.enableCompression = b
	}
}

func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *cfg) {
		cfg.handshakeTimeout = d
	}
}
