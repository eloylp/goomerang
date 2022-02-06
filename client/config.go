package client

import (
	"crypto/tls"
)

type Option func(cfg *Config)

type Config struct {
	TargetServer    string
	OnCloseHook     func()
	OnErrorHook     func(err error)
	TLSConfig       *tls.Config
	MaxConcurrency  int
	ReadBufferSize  int
	WriteBufferSize int
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook:    func(err error) {},
		OnCloseHook:    func() {},
		MaxConcurrency: 10,
	}
	return cfg
}

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
