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
	SendSyncResponseTime  prometheus.Histogram
	CurrentStatus         prometheus.Gauge
	Errors                prometheus.Counter
)

func init() {
	Configure(Config{
		MessageInflightTimeBuckets:   prometheus.DefBuckets,
		ReceivedMessageSizeBuckets:   prometheus.DefBuckets,
		SentMessageSizeBuckets:       prometheus.DefBuckets,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
		SendSyncResponseTimeBuckets:  prometheus.DefBuckets,
	})
}

func Configure(config Config) {

	MessageInflightTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "received_message_inflight_duration_seconds",
		Help:      "The time the message spent over the wire till received",
		Buckets:   config.MessageInflightTimeBuckets,
	})

	ReceivedMessageSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "received_message_size_bytes",
		Help:      "The size of the received messages in bytes",
		Buckets:   config.ReceivedMessageSizeBuckets,
	})

	SentMessageSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "sent_message_size_bytes",
		Help:      "The size of the sent messages in bytes",
		Buckets:   config.SentMessageSizeBuckets,
	})

	MessageProcessingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_processing_duration_seconds",
		Help:      "The time spent in message handler execution",
		Buckets:   config.MessageProcessingTimeBuckets,
	})

	SendSyncResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_sync_send_duration_seconds",
		Help:      "The time spent in a synchronous message sending operation",
		Buckets:   config.SendSyncResponseTimeBuckets,
	})

	CurrentStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "status",
		Help:      "The current status of the client (0 => New, 1 => Running, 2=> Closing, 3 => closed)",
	})

	Errors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "errors_total",
		Help:      "The errors happened in client",
	})
}

type Config struct {
	MessageInflightTimeBuckets   []float64
	ReceivedMessageSizeBuckets   []float64
	SentMessageSizeBuckets       []float64
	MessageProcessingTimeBuckets []float64
	SendSyncResponseTimeBuckets  []float64
}
