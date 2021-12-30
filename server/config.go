package server

import (
	"time"
)

type Option func(cfg *Config)

type timedHook func(name string, duration time.Duration)

func WithListenAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.ListenURL = addr
	}
}

func WithErrorHook(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.ErrorHook = h
	}
}

func WithOnCloseHook(h func()) Option {
	return func(cfg *Config) {
		cfg.OnCloseHook = h
	}
}

func WithOnMessageProcessedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageProcessedHook = h
	}
}

type Config struct {
	ListenURL              string
	ErrorHook              func(err error)
	OnCloseHook            func()
	OnMessageProcessedHook timedHook
}

func defaultConfig() *Config {
	cfg := &Config{
		ErrorHook:              func(err error) {},
		OnCloseHook:            func() {},
		OnMessageProcessedHook: func(name string, duration time.Duration) {},
	}
	return cfg
}
