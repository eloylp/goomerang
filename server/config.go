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

type Config struct {
	ListenURL    string
	ErrorHandler func(err error)
}
