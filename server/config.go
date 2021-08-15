package server

type Option func(cfg *Config)

func WithListenAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.ListenURL = addr
	}
}

func WithErrorHandler(h func(err error)) Option {
	return func(cfg *Config) {
		cfg.ErrorHandler = h
	}
}

func WithOnCloseHandler(h func()) Option {
	return func(cfg *Config) {
		cfg.OnCloseHandler = h
	}
}

type Config struct {
	ListenURL      string
	ErrorHandler   func(err error)
	OnCloseHandler func()
}

func defaultConfig() *Config {
	cfg := &Config{
		ErrorHandler:   func(err error) {},
		OnCloseHandler: func() {},
	}
	return cfg
}
