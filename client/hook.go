package client

import (
	"go.eloylp.dev/goomerang/internal/config"
)

type hooks struct {
	config.Hooks
	onConfiguration []func(cfg *Cfg)
}

func (h *hooks) AppendOnConfiguration(f func(cfg *Cfg)) {
	h.onConfiguration = append(h.onConfiguration, f)
}

func (h *hooks) ExecOnConfiguration(cfg *Cfg) {
	for _, f := range h.onConfiguration {
		f(cfg)
	}
}
