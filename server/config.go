package server

import (
	"crypto/tls"
	"log"
	"time"
)

type Config struct {
	ListenURL             string
	OnErrorHook           func(err error)
	OnCloseHook           func()
	TLSConfig             *tls.Config
	MaxConcurrency        int
	ReadBufferSize        int
	WriteBufferSize       int
	HTTPWriteTimeout      time.Duration
	HTTPReadTimeout       time.Duration
	HTTPReadHeaderTimeout time.Duration
	HandshakeTimeout      time.Duration
	EnableCompression     bool
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook: func(err error) {
			log.Printf("goomerang error: %v", err) //nolint:forbidigo
		},
		OnCloseHook:    func() {},
		MaxConcurrency: 10,
	}
	return cfg
}

func endpoint(cfg *Config) string {
	if cfg.TLSConfig != nil {
		return "/wss"
	}
	return "/ws"
}
