package server

import (
	"crypto/tls"
	"time"

	"go.eloylp.dev/goomerang/internal/config"
)

type cfg struct {
	listenURL             string
	hooks                 *config.Hooks
	tlsConfig             *tls.Config
	maxConcurrency        int
	readBufferSize        int
	writeBufferSize       int
	httpWriteTimeout      time.Duration
	httpReadTimeout       time.Duration
	httpReadHeaderTimeout time.Duration
	handshakeTimeout      time.Duration
	enableCompression     bool
}

func defaultConfig() *cfg {
	cfg := &cfg{
		hooks:          &config.Hooks{},
		maxConcurrency: 10,
	}
	return cfg
}

func endpoint(cfg *cfg) string {
	if cfg.tlsConfig != nil {
		return "/wss"
	}
	return "/ws"
}
