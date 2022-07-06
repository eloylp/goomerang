//go:build unit

package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnConfigurationHook(t *testing.T) {
	defCfg := defaultConfig()
	hooks := &hooks{}
	hooks.AppendOnConfiguration(func(cfg *Cfg) {
		cfg.EnableCompression = true
	})
	hooks.AppendOnConfiguration(func(cfg *Cfg) {
		cfg.HeartbeatInterval = 1000
	})
	hooks.ExecOnConfiguration(defCfg)
	assert.True(t, defCfg.EnableCompression)
	assert.Equal(t, time.Duration(1000), defCfg.HeartbeatInterval)
}
