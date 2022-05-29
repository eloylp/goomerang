package client

import (
	"crypto/tls"
	"net/url"
	"time"
)

type Cfg struct {
	TargetServer      string
	hooks             *hooks
	TLSConfig         *tls.Config
	MaxConcurrency    int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	EnableCompression bool
	HeartbeatInterval time.Duration
}

func defaultConfig() *Cfg {
	cfg := &Cfg{
		hooks:             &hooks{},
		HeartbeatInterval: 5 * time.Second,
		MaxConcurrency:    10,
	}
	return cfg
}

func serverURL(cfg *Cfg) url.URL {
	if cfg.TLSConfig != nil {
		return url.URL{Scheme: "wss", Host: cfg.TargetServer, Path: "/wss"}
	}
	return url.URL{Scheme: "ws", Host: cfg.TargetServer, Path: "/ws"}
}
