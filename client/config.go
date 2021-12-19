package client

import (
	"time"
)

type Option func(cfg *Config)

type timedHandler func(name string, duration time.Duration)

type Config struct {
	TargetServer              string
	OnCloseHandler            func()
	OnErrorHandler            func(err error)
	OnMessageProcessedHandler timedHandler
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHandler: func(err error) {},
		OnCloseHandler: func() {},
	}
	return cfg
}

func WithTargetServer(addr string) Option {
	return func(cfg *Config) {
		cfg.TargetServer = addr
	}
}

func WithOnCloseHandler(h func()) Option {
	return func(cfg *Config) {
		cfg.OnCloseHandler = h
	}
}

func WithOnErrorHandler(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.OnErrorHandler = h
	}
}

func WithOnMessageProcessedHandler(h timedHandler) Option {
	return func(cfg *Config) {
		cfg.OnMessageProcessedHandler = h
	}
}
