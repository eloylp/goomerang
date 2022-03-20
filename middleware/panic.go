package middleware

import (
	"go.eloylp.dev/goomerang/message"
)

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
