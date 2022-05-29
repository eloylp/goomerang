package config

type Hooks struct {
	onStatusChange []func(status uint32)
	onClose        []func()
	onError        []func(err error)
}

func (h *Hooks) AppendOnStatusChange(f func(status uint32)) {
	h.onStatusChange = append(h.onStatusChange, f)
}

func (h *Hooks) AppendOnClose(f func()) {
	h.onClose = append(h.onClose, f)
}

func (h *Hooks) AppendOnError(f func(err error)) {
	h.onError = append(h.onError, f)
}

func (h *Hooks) ExecOnStatusChange(status uint32) {
	for _, f := range h.onStatusChange {
		f(status)
	}
}

func (h *Hooks) ExecOnclose() {
	for _, f := range h.onClose {
		f()
	}
}

func (h *Hooks) ExecOnError(err error) {
	for _, f := range h.onError {
		f(err)
	}
}
