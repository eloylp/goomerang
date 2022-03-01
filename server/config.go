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
}

func defaultConfig() *Config {
	cfg := &Config{
		OnErrorHook: func(err error) {
			log.Printf("goomerang error: %v", err) //nolint:forbidigo
		},
		MaxConcurrency: 10,
	}
	return cfg
}
