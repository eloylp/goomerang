package server

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Config)

func WithListenAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.ListenURL = addr
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

func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(cfg *Config) {
		cfg.TLSConfig = tlsConfig
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

func WithHTTPWriteTimeout(t time.Duration) Option {
	return func(cfg *Config) {
		cfg.HTTPWriteTimeout = t
	}
}

func WithHTTPReadTimeout(t time.Duration) Option {
	return func(cfg *Config) {
		cfg.HTTPReadTimeout = t
	}
}

func WithHTTPReadHeaderTimeout(t time.Duration) Option {
	return func(cfg *Config) {
		cfg.HTTPReadHeaderTimeout = t
	}
}
