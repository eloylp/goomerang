package config

type Hooks struct {
	onStatusChange []func(status uint32)
	onClose        []func()
	onError        []func(err error)
	onHandlerStart []func(kind string)
	onHandlerEnd   []func(kind string)
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

func (h *Hooks) AppendOnHandlerStart(f func(kind string)) {
	h.onHandlerStart = append(h.onHandlerStart, f)
}

func (h *Hooks) AppendOnHandlerEnd(f func(kind string)) {
	h.onHandlerEnd = append(h.onHandlerEnd, f)
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

func (h *Hooks) ExecOnHandlerStart(kind string) {
	for _, f := range h.onHandlerStart {
		f(kind)
	}
}

func (h *Hooks) ExecOnHandlerEnd(kind string) {
	for _, f := range h.onHandlerEnd {
		f(kind)
	}
}
