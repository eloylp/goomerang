package goomerang

type ServerOption func(cfg *ServerConfig)
type ClientOption func(cfg *ClientConfig)

func WithListenAddr(addr string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ListenURL = addr
	}
}

type ServerConfig struct {
	ListenURL string
}

type ClientConfig struct {
	TargetServer string
}

func WithTargetServer(addr string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.TargetServer = addr
	}
}
