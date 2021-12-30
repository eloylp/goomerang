package client

import (
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
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook: func(err error) {},
		OnCloseHook: func() {},
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

func WithOnMessageProcessedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageProcessedHook = h
	}
}

func WithOnMessageReceivedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageReceivedHook = h
	}
}
