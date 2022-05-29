package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnConfigurationHook(t *testing.T) {
	defCfg := defaultConfig()
	hooks := &hooks{}
	hooks.AppendOnConfiguration(func(cfg *Cfg) {
		cfg.EnableCompression = true
	})
	hooks.AppendOnConfiguration(func(cfg *Cfg) {
		cfg.WriteBufferSize = 1000
	})
	hooks.ExecOnConfiguration(defCfg)
	assert.True(t, defCfg.EnableCompression)
	assert.Equal(t, 1000, defCfg.WriteBufferSize)
}
