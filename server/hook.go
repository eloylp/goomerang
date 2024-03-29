package server

import (
	"time"

	"go.eloylp.dev/goomerang/internal/config"
)

type hooks struct {
	config.Hooks
	onConfiguration   []func(cfg *Cfg)
	onBroadcast       []func(fqdn string, result []BroadcastResult, duration time.Duration)
	onClientBroadcast []func(fqdn string)
	onSubscribe       []func(topic string)
	onPublish         []func(topic, fqdn string)
	onUnsubscribe     []func(topic string)
}

func (h *hooks) AppendOnConfiguration(f func(cfg *Cfg)) {
	h.onConfiguration = append(h.onConfiguration, f)
}

func (h *hooks) ExecOnConfiguration(cfg *Cfg) {
	for _, f := range h.onConfiguration {
		f(cfg)
	}
}

func (h *hooks) AppendOnSubscribe(f func(topic string)) {
	h.onSubscribe = append(h.onSubscribe, f)
}

func (h *hooks) ExecOnSubscribe(topic string) {
	for _, f := range h.onSubscribe {
		f(topic)
	}
}

func (h *hooks) AppendOnBroadcast(f func(fqdn string, result []BroadcastResult, duration time.Duration)) {
	h.onBroadcast = append(h.onBroadcast, f)
}

func (h *hooks) ExecOnBroadcast(fqdn string, result []BroadcastResult, duration time.Duration) {
	for _, f := range h.onBroadcast {
		f(fqdn, result, duration)
	}
}

func (h *hooks) AppendOnClientBroadcast(f func(fqdn string)) {
	h.onClientBroadcast = append(h.onClientBroadcast, f)
}

func (h *hooks) ExecOnClientBroadcast(fqdn string) {
	for _, f := range h.onClientBroadcast {
		f(fqdn)
	}
}

func (h *hooks) AppendOnPublish(f func(topic, fqdn string)) {
	h.onPublish = append(h.onPublish, f)
}

func (h *hooks) ExecOnPublish(topic, fqdn string) {
	for _, f := range h.onPublish {
		f(topic, fqdn)
	}
}

func (h *hooks) AppendOnUnsubscribe(f func(topic string)) {
	h.onUnsubscribe = append(h.onUnsubscribe, f)
}

func (h *hooks) ExecOnUnsubscribe(topic string) {
	for _, f := range h.onUnsubscribe {
		f(topic)
	}
}
