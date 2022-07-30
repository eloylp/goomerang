//go:build unit

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

func TestOnSubscribeHook(t *testing.T) {
	hooks := &hooks{}
	hooks.AppendOnSubscribe(func(topic string) {
		assert.Equal(t, "topic.a", topic)
	})
	hooks.AppendOnSubscribe(func(topic string) {
		assert.Equal(t, "topic.a", topic)
	})
	hooks.ExecOnSubscribe("topic.a")
}

func TestOnPublishHook(t *testing.T) {
	hooks := &hooks{}
	hooks.AppendOnPublish(func(topic, fqdn string) {
		assert.Equal(t, "topic.a", topic)
		assert.Equal(t, "message.a", fqdn)
	})
	hooks.AppendOnPublish(func(topic, fqdn string) {
		assert.Equal(t, "topic.a", topic)
		assert.Equal(t, "message.a", fqdn)
	})
	hooks.ExecOnPublish("topic.a", "message.a")
}

func TestOnUnsubscribeHook(t *testing.T) {
	hooks := &hooks{}
	hooks.AppendOnUnsubscribe(func(topic string) {
		assert.Equal(t, "topic.a", topic)
	})
	hooks.AppendOnUnsubscribe(func(topic string) {
		assert.Equal(t, "topic.a", topic)
	})
	hooks.ExecOnUnsubscribe("topic.a")
}
