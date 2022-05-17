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
	CurrentStatus         prometheus.Gauge
	Errors                prometheus.Counter
)

func init() {
	Configure(Config{
		MessageInflightTimeBuckets:   prometheus.DefBuckets,
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
		Buckets:   c.MessageInflightTimeBuckets,
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

	CurrentStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "status",
		Help:      "The current status of the server (0 => New, 1 => Running, 2=> Closing, 3 => closed)",
	})

	Errors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "errors_total",
		Help:      "The errors happened in server",
	})
}

type Config struct {
	MessageInflightTimeBuckets   []float64
	ReceivedMessageSizeBuckets   []float64
	SentMessageSizeBuckets       []float64
	MessageProcessingTimeBuckets []float64
	BroadcastSentTimeBuckets     []float64
}
