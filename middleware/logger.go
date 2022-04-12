package middleware

import (
	"time"

	"github.com/sirupsen/logrus"

	"go.eloylp.dev/goomerang/message"
)

func Logger(logger logrus.FieldLogger) message.Middleware {
	return func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			sender := NewSender(s)
			start := time.Now()
			h.Handle(sender, msg)
			duration := time.Since(start)
			logger.WithFields(logrus.Fields{
				"type":           msg.Metadata.Type,
				"metadata":       msg.Metadata,
				"headers":        msg.Header,
				"processingTime": duration,
				"payload":        msg.Payload,
			}).Debug("message processed")
		})
	}
}
