package server

type Option func(cfg *Config)

func WithListenAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.ListenURL = addr
	}
}

type Config struct {
	ListenURL string
}
