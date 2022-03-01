package client

import (
	"crypto/tls"
)

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
