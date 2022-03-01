package client

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Config)

func WithTargetServer(addr string) Option {
	return func(cfg *Config) {
		cfg.TargetServer = addr
	}
}

func WithOnCloseHook(h func()) Option {
	return func(cfg *Config) {
		cfg.OnCloseHook = h
	}
}

func WithOnErrorHook(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.OnErrorHook = h
	}
}

func WithMaxConcurrency(n int) Option {
	return func(cfg *Config) {
		cfg.MaxConcurrency = n
	}
}

func WithWithTLSConfig(tlsCfg *tls.Config) Option {
	return func(cfg *Config) {
		cfg.TLSConfig = tlsCfg
	}
}

func WithReadBufferSize(s int) Option {
	return func(cfg *Config) {
		cfg.ReadBufferSize = s
	}
}

func WithWriteBufferSize(s int) Option {
	return func(cfg *Config) {
		cfg.WriteBufferSize = s
	}
}

func WithCompressionEnabled(b bool) Option {
	return func(cfg *Config) {
		cfg.EnableCompression = b
	}
}

func WithHandShakeTimeout(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.HandshakeTimeout = d
	}
}
