package server

import (
	"crypto/tls"
	"log"
	"time"
)

type Option func(cfg *Config)

type timedHook func(name string, duration time.Duration)

type Config struct {
	ListenURL              string
	OnErrorHook            func(err error)
	OnCloseHook            func()
	OnMessageProcessedHook timedHook
	OnMessageReceivedHook  timedHook
	TLSConfig              *tls.Config
	MaxConcurrency         int
	ReadBufferSize         int
	WriteBufferSize        int
	HTTPWriteTimeout       time.Duration
	HTTPReadTimeout        time.Duration
	HTTPReadHeaderTimeout  time.Duration
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook: func(err error) {
			log.Printf("goomerang error: %v", err) //nolint:forbidigo
		},
		MaxConcurrency: 10,
	}
	return cfg
}

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

func WithOnMessageReceivedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageReceivedHook = h
	}
}

func WithOnMessageProcessedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageProcessedHook = h
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
