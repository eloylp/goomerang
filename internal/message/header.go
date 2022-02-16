package message

type Header map[string]string

func (h Header) Get(key string) string {
	return h[key]
}

func (h Header) Add(key, value string) {
	h[key] = value
}
