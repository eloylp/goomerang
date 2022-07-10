package middleware

import (
	"go.eloylp.dev/goomerang/message"
)

// Panic stops panic propagation and executes the
// provided function. The panic value will be passed to it.
func Panic(panicHandler func(p interface{})) message.Middleware {
	return func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
			defer func() {
				if p := recover(); p != nil {
					panicHandler(p)
				}
			}()
			h.Handle(sender, msg)
		})
	}
}
