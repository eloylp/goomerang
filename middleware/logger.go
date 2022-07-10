package middleware

import (
	"time"

	"github.com/sirupsen/logrus"

	"go.eloylp.dev/goomerang/message"
)

// Logger allows passing a customized logrus logger. It will
// log the main attributes of all messages.
//
// It will only log messages in Debug level.
func Logger(logger logrus.FieldLogger) message.Middleware {
	return func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			sender := NewSender(s)
			start := time.Now()
			h.Handle(sender, msg)
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"kind":           msg.Metadata.Kind,
				"metadata":       msg.Metadata,
				"headers":        msg.Header,
				"processingTime": duration,
				"payload":        msg.Payload,
			}).Debug("message processed")
		})
	}
}
