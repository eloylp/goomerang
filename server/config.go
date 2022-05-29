package server

import (
	"crypto/tls"
	"time"
)

type Cfg struct {
	ListenURL             string
	hooks                 *hooks
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

func defaultConfig() *Cfg {
	cfg := &Cfg{
		hooks:          &hooks{},
		MaxConcurrency: 10,
	}
	return cfg
}

func endpoint(cfg *Cfg) string {
	if cfg.TLSConfig != nil {
		return "/wss"
	}
	return "/ws"
}
