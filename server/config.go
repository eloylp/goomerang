package server

import (
	"crypto/tls"
	"time"

	"go.eloylp.dev/goomerang/internal/config"
)

type Config struct {
	ListenURL             string
	Hooks                 *config.Hooks
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
		Hooks:          &config.Hooks{},
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
