package client

import (
	"crypto/tls"
	"net/url"
	"time"

	"go.eloylp.dev/goomerang/internal/config"
)

type cfg struct {
	targetServer      string
	hooks             *config.Hooks
	tlsConfig         *tls.Config
	maxConcurrency    int
	readBufferSize    int
	writeBufferSize   int
	handshakeTimeout  time.Duration
	enableCompression bool
	heartbeatInterval time.Duration
}

func defaultConfig() *cfg {
	cfg := &cfg{
		hooks:             &config.Hooks{},
		heartbeatInterval: 5 * time.Second,
		maxConcurrency:    10,
	}
	return cfg
}

func serverURL(cfg *cfg) url.URL {
	if cfg.tlsConfig != nil {
		return url.URL{Scheme: "wss", Host: cfg.targetServer, Path: "/wss"}
	}
	return url.URL{Scheme: "ws", Host: cfg.targetServer, Path: "/ws"}
}
