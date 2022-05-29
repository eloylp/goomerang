package client

import (
	"crypto/tls"
	"net/url"
	"time"

	"go.eloylp.dev/goomerang/internal/config"
)

type Config struct {
	TargetServer      string
	Hooks             *config.Hooks
	TLSConfig         *tls.Config
	MaxConcurrency    int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	EnableCompression bool
	HeartbeatInterval time.Duration
}

func defaultConfig() *Config {
	cfg := &Config{
		Hooks:             &config.Hooks{},
		HeartbeatInterval: 5 * time.Second,
		MaxConcurrency:    10,
	}
	return cfg
}

func serverURL(cfg *Config) url.URL {
	if cfg.TLSConfig != nil {
		return url.URL{Scheme: "wss", Host: cfg.TargetServer, Path: "/wss"}
	}
	return url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
}
