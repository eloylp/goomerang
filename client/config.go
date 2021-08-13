package client

type Option func(cfg *Config)

type Config struct {
	TargetServer   string
	OnCloseHandler func()
	OnErrorHandler func(err error)
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

func WithOnErrorHandler(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.OnErrorHandler = h
	}
}
