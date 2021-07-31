package client

type Option func(cfg *Config)

type Config struct {
	TargetServer   string
	OnCloseHandler func()
}

func WithTargetServer(addr string) Option {
	return func(cfg *Config) {
		cfg.TargetServer = addr
	}
}

func WithOnCloseHandler(h func()) Option {
	return func(cfg *Config) {
		cfg.OnCloseHandler = h
	}
}
