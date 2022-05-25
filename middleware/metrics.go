package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/metrics/client"
	server "go.eloylp.dev/goomerang/metrics/server"
)

func PromHistograms(c PromConfig) (message.Middleware, error) {

	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("metrics middleware config: validation error: %w", err)
	}
	return func(h message.Handler) message.Handler {
		return message.HandlerFunc(func(s message.Sender, msg *message.Message) {
			c.MessageInflightTime.WithLabelValues(msg.Metadata.Type).Observe(time.Since(msg.Metadata.Creation).Seconds())
			c.MessageReceivedSize.WithLabelValues(msg.Metadata.Type).Observe(float64(msg.Metadata.PayloadSize))

			wrappedSender := NewMeteredSender(s, c)
			start := time.Now()
			h.Handle(wrappedSender, msg)

			sentMessage := wrappedSender.Msg()
			if sentMessage != nil {
				c.MessageProcessingTime.WithLabelValues(sentMessage.Metadata.Type).Observe(float64(time.Since(start)))
				c.MessageSentSize.WithLabelValues(sentMessage.Metadata.Type).Observe(float64(wrappedSender.Bytes()))
			}
		})
	}, nil
}

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

func ClientStatusMetricHook(status uint32) {
	client.CurrentStatus.Set(float64(status))
}

func ServerStatusMetricHook(status uint32) {
	server.CurrentStatus.Set(float64(status))
}
