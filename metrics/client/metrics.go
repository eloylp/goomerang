package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessageInflightTime         *prometheus.HistogramVec
	MessageReceivedSize         *prometheus.HistogramVec
	MessageSentSize             *prometheus.HistogramVec
	MessageProcessingTime       *prometheus.HistogramVec
	MessageSentSyncResponseTime *prometheus.HistogramVec
	MessageSentTime             *prometheus.HistogramVec
	CurrentStatus               prometheus.Gauge
	ConcurrentWorkers           prometheus.Gauge
	ConfigMaxConcurrency        prometheus.Gauge
	Errors                      prometheus.Counter
)

func init() {
	Configure(Config{
		MessageInflightTimeBuckets:   prometheus.DefBuckets,
		ReceivedMessageSizeBuckets:   prometheus.DefBuckets,
		SentMessageSizeBuckets:       prometheus.DefBuckets,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
		SendSyncResponseTimeBuckets:  prometheus.DefBuckets,
		SendTimeBuckets:              prometheus.DefBuckets,
	})
}

func Configure(config Config) {

	MessageInflightTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_received_inflight_duration_seconds",
		Help:      "The time the message spent over the wire till received",
		Buckets:   config.MessageInflightTimeBuckets,
	}, []string{"type"})

	MessageReceivedSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_received_size_bytes",
		Help:      "The size of the received messages in bytes",
		Buckets:   config.ReceivedMessageSizeBuckets,
	}, []string{"type"})

	MessageSentSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_sent_size_bytes",
		Help:      "The size of the sent messages in bytes",
		Buckets:   config.SentMessageSizeBuckets,
	}, []string{"type"})

	MessageSentTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_sent_duration_seconds",
		Help:      "The time spent in during asynchronous message sending operation (buffer)",
		Buckets:   config.SendTimeBuckets,
	}, []string{"type"})

	MessageProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_processing_duration_seconds",
		Help:      "The time spent in message handler execution",
		Buckets:   config.MessageProcessingTimeBuckets,
	}, []string{"type"})

	MessageSentSyncResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "message_sync_sent_duration_seconds",
		Help:      "The time spent in a synchronous message sending operation",
		Buckets:   config.SendSyncResponseTimeBuckets,
	}, []string{"type"})

	CurrentStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "status",
		Help:      "The current status of the client (0 => New, 1 => Running, 2=> Closing, 3 => closed)",
	})

	ConcurrentWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "concurrent_workers",
		Help:      "The current number of running workers in the client",
	})

	ConfigMaxConcurrency = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "client",
		Name:      "config_max_concurrency",
		Help:      "The configured maximum number of parallel workers in the client",
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
	SendTimeBuckets              []float64
}
