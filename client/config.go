package client

import (
	"crypto/tls"
	"net/url"
	"time"
)

type Config struct {
	TargetServer      string
	OnCloseHook       func()
	OnErrorHook       func(err error)
	TLSConfig         *tls.Config
	MaxConcurrency    int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	EnableCompression bool
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook:    func(err error) {},
		OnCloseHook:    func() {},
		MaxConcurrency: 10,
	}
	return cfg
}

func serverURL(cfg *Config) url.URL {
	if cfg.TLSConfig != nil {
		return url.URL{Scheme: "wss", Host: cfg.TargetServer, Path: "/wss"}
	}
	return url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
}
