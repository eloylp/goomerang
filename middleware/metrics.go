package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"go.eloylp.dev/goomerang/message"
)

// PromHistograms will return a middleware ready to register
// Prometheus metrics. It accepts a PromConfig type parameter
// for configuring the expected histograms.
func PromHistograms(c PromConfig) (message.Middleware, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("metrics middleware config: validation error: %w", err)
	}
	return func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			c.MessageInflightTime.WithLabelValues(msg.Metadata.Kind).Observe(time.Since(msg.Metadata.Creation).Seconds())
			c.MessageReceivedSize.WithLabelValues(msg.Metadata.Kind).Observe(float64(msg.Metadata.PayloadSize))

			wrappedSender := NewMeteredSender(s, c)
			start := time.Now()
			h.Handle(wrappedSender, msg)

			c.MessageProcessingTime.WithLabelValues(msg.Metadata.Kind).Observe(time.Since(start).Seconds())
			sentMessage := wrappedSender.Msg()
			if sentMessage != nil {
				c.MessageSentSize.WithLabelValues(sentMessage.Metadata.Kind).Observe(float64(wrappedSender.Bytes()))
			}
		})
	}, nil
}

// PromConfig holds all needed histograms for registering
// message time and sizes.
type PromConfig struct {
	MessageInflightTime   *prometheus.HistogramVec
	MessageReceivedSize   *prometheus.HistogramVec
	MessageProcessingTime *prometheus.HistogramVec
	MessageSentSize       *prometheus.HistogramVec
	MessageSentTime       *prometheus.HistogramVec
}

func (c PromConfig) Validate() error {
	if c.MessageInflightTime == nil {
		return errors.New("validate: messageInflightTime be non nil")
	}
	if c.MessageReceivedSize == nil {
		return errors.New("validate: receivedMessageSize be non nil")
	}
	if c.MessageProcessingTime == nil {
		return errors.New("validate: MessageProcessingTime be non nil")
	}
	if c.MessageSentSize == nil {
		return errors.New("validate: MessageSentSize be non nil")
	}
	if c.MessageSentTime == nil {
		return errors.New("validate: MessageSentTime be non nil")
	}
	return nil
}
