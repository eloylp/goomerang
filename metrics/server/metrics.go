package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessageInflightTime   *prometheus.HistogramVec
	ReceivedMessageSize   *prometheus.HistogramVec
	SentMessageSize       *prometheus.HistogramVec
	SentMessageTime       *prometheus.HistogramVec
	MessageProcessingTime *prometheus.HistogramVec
	BroadcastSentTime     *prometheus.HistogramVec
	CurrentStatus         prometheus.Gauge
	Errors                prometheus.Counter
)

func init() {
	Configure(Config{
		MessageInflightTimeBuckets:   prometheus.DefBuckets,
		ReceivedMessageSizeBuckets:   prometheus.DefBuckets,
		SentMessageSizeBuckets:       prometheus.DefBuckets,
		SentMessageTimeBuckets:       prometheus.DefBuckets,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
		BroadcastSentTimeBuckets:     prometheus.DefBuckets,
	})
}

func Configure(c Config) {

	MessageInflightTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_received_inflight_duration_seconds",
		Help:      "The time the message spent over the wire till received",
		Buckets:   c.MessageInflightTimeBuckets,
	}, []string{"type"})

	ReceivedMessageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_received_size_bytes",
		Help:      "The size of the received messages in bytes",
		Buckets:   c.ReceivedMessageSizeBuckets,
	}, []string{"type"})

	SentMessageSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_sent_size_bytes",
		Help:      "The size of the sent messages in bytes",
		Buckets:   c.SentMessageSizeBuckets,
	}, []string{"type"})

	SentMessageTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_sent_duration_seconds",
		Help:      "The time spent in during asynchronous message sending operation (buffer)",
		Buckets:   c.SentMessageTimeBuckets,
	}, []string{"type"})

	MessageProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_processing_duration_seconds",
		Help:      "The time spent in message handler execution",
		Buckets:   c.MessageProcessingTimeBuckets,
	}, []string{"type"})

	BroadcastSentTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_broadcast_sent_duration_seconds",
		Help:      "The time spent in a broadcast operation",
		Buckets:   c.BroadcastSentTimeBuckets,
	}, []string{"type"})

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
	SentMessageTimeBuckets       []float64
}
