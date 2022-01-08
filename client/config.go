package client

import (
	"crypto/tls"
	"time"
)

type Option func(cfg *Config)

type timedHook func(name string, duration time.Duration)

type Config struct {
	TargetServer           string
	OnCloseHook            func()
	OnErrorHook            func(err error)
	OnMessageProcessedHook timedHook
	OnMessageReceivedHook  timedHook
	TLSConfig              *tls.Config
	MaxConcurrency         int
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

func WithWithTLSConfig(tlsCfg *tls.Config) Option {
	return func(cfg *Config) {
		cfg.TLSConfig = tlsCfg
	}
}
