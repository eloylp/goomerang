package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ReceivedMessageSize   prometheus.Histogram
	SentMessageSize       prometheus.Histogram
	MessageProcessingTime prometheus.Histogram
)

func init() {
	Configure(Config{
		ReceivedMessageSizeBuckets:   prometheus.DefBuckets,
		SentMessageSizeBuckets:       prometheus.DefBuckets,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
	})
}

func Configure(c Config) {
	ReceivedMessageSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "received_message_size_bytes",
		Help:      "The size of the received messages in bytes",
		Buckets:   c.ReceivedMessageSizeBuckets,
	})

	SentMessageSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "sent_message_size_bytes",
		Help:      "The size of the sent messages in bytes",
		Buckets:   c.SentMessageSizeBuckets,
	})

	MessageProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_processing_duration_seconds",
		Help:      "The time spent in message handler execution",
		Buckets:   c.MessageProcessingTimeBuckets,
	})

}

type Config struct {
	ReceivedMessageSizeBuckets   []float64
	SentMessageSizeBuckets       []float64
	MessageProcessingTimeBuckets []float64
}
