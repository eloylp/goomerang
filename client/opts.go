package client

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Cfg)

func WithTargetServer(addr string) Option {
	return func(cfg *Cfg) {
		cfg.TargetServer = addr
	}
}

func WithOnStatusChangeHook(h func(status uint32)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnStatusChange(h)
	}
}

func WithHeartbeatInterval(interval time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HeartbeatInterval = interval
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

func WithOnHandlerStart(h func(kind string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnHandlerStart(h)
	}
}

func WithOnHandlerEnd(h func(kind string)) Option {
	return func(cfg *Cfg) {
		cfg.hooks.AppendOnHandlerEnd(h)
	}
}

func WithMaxConcurrency(n int) Option {
	return func(cfg *Cfg) {
		cfg.MaxConcurrency = n
	}
}

func WithWithTLSConfig(tlsCfg *tls.Config) Option {
	return func(cfg *Cfg) {
		cfg.TLSConfig = tlsCfg
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

func WithCompressionEnabled(b bool) Option {
	return func(cfg *Cfg) {
		cfg.EnableCompression = b
	}
}

func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *Cfg) {
		cfg.HandshakeTimeout = d
	}
}
