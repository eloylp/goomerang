package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessageInflightTime   prometheus.Histogram
	ReceivedMessageSize   prometheus.Histogram
	SentMessageSize       prometheus.Histogram
	MessageProcessingTime prometheus.Histogram
	BroadcastSentTime     prometheus.Histogram
)

func init() {
	Configure(Config{
		MessageInflightTime:          prometheus.DefBuckets,
		ReceivedMessageSizeBuckets:   prometheus.DefBuckets,
		SentMessageSizeBuckets:       prometheus.DefBuckets,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
		BroadcastSentTimeBuckets:     prometheus.DefBuckets,
	})
}

func Configure(c Config) {

	MessageInflightTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "received_message_inflight_duration_seconds",
		Help:      "The time the message spent over the wire till received",
		Buckets:   c.ReceivedMessageSizeBuckets,
	})

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

	BroadcastSentTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_broadcast_sent_duration_seconds",
		Help:      "The time spent in a broadcast operation",
		Buckets:   c.BroadcastSentTimeBuckets,
	})

}

type Config struct {
	MessageInflightTime          []float64
	ReceivedMessageSizeBuckets   []float64
	SentMessageSizeBuckets       []float64
	MessageProcessingTimeBuckets []float64
	BroadcastSentTimeBuckets     []float64
}
