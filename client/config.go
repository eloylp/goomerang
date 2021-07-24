package client

type Option func(cfg *Config)

type Config struct {
	TargetServer string
}

func WithTargetServer(addr string) Option {
	return func(cfg *Config) {
		cfg.TargetServer = addr
	}
}
