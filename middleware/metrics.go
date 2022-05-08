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
			c.MessageInflightTime.Observe(time.Since(msg.Metadata.Creation).Seconds())
			c.ReceivedMessageSize.Observe(float64(msg.Metadata.PayloadSize))
			wrappedSender := NewSender(s)
			start := time.Now()
			h.Handle(wrappedSender, msg)
			c.MessageProcessingTime.Observe(float64(time.Since(start)))
			c.SentMessageSize.Observe(float64(wrappedSender.Bytes()))
		})
	}, nil
}

type PromConfig struct {
	MessageInflightTime   prometheus.Histogram
	ReceivedMessageSize   prometheus.Histogram
	MessageProcessingTime prometheus.Histogram
	SentMessageSize       prometheus.Histogram
}

func (c PromConfig) Validate() error {
	if c.MessageInflightTime == nil {
		return errors.New("validate: messageInflightTime be non nil")
	}
	if c.ReceivedMessageSize == nil {
		return errors.New("validate: receivedMessageSize be non nil")
	}
	if c.MessageProcessingTime == nil {
		return errors.New("validate: MessageProcessingTime be non nil")
	}
	if c.SentMessageSize == nil {
		return errors.New("validate: SentMessageSize be non nil")
	}
	return nil
}

func ClientStatusMetricHook(status uint32) {
	client.CurrentStatus.Set(float64(status))
}

func ServerStatusMetricHook(status uint32) {
	server.CurrentStatus.Set(float64(status))
}
