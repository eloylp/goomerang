package server

import (
	"log"
	"time"
)

type Option func(cfg *Config)

type timedHook func(name string, duration time.Duration)

func WithListenAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.ListenURL = addr
	}
}

func WithOnErrorHook(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.OnErrorHook = h
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

func WithOnMessageReceivedHook(h timedHook) Option {
	return func(cfg *Config) {
		cfg.OnMessageReceivedHook = h
	}
}

type Config struct {
	ListenURL              string
	OnErrorHook            func(err error)
	OnCloseHook            func()
	OnMessageProcessedHook timedHook
	OnMessageReceivedHook  timedHook
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook: func(err error) {
			log.Printf("goomerang error: %v", err)
		},
	}
	return cfg
}
