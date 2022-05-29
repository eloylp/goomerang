package client

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessageInflightTime         *prometheus.HistogramVec
	MessageReceivedSize         *prometheus.HistogramVec
	MessageSentSize             *prometheus.HistogramVec
	MessageSentTime             *prometheus.HistogramVec
	MessageProcessingTime       *prometheus.HistogramVec
	MessageBroadcastSentTime    *prometheus.HistogramVec
	CurrentStatus               prometheus.Gauge
	ConcurrentHandlers          *prometheus.GaugeVec
	ConfigMaxConcurrentHandlers prometheus.Gauge
	Errors                      prometheus.Counter
)

func init() {
	Configure(Config{
		MessageInflightTimeBuckets:      prometheus.DefBuckets,
		MessageReceivedSizeBuckets:      prometheus.DefBuckets,
		MessageSentSizeBuckets:          prometheus.DefBuckets,
		MessageSentTimeBuckets:          prometheus.DefBuckets,
		MessageProcessingTimeBuckets:    prometheus.DefBuckets,
		MessageBroadcastSentTimeBuckets: prometheus.DefBuckets,
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

	MessageReceivedSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_received_size_bytes",
		Help:      "The size of the received messages in bytes",
		Buckets:   c.MessageReceivedSizeBuckets,
	}, []string{"type"})

	MessageSentSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_sent_size_bytes",
		Help:      "The size of the sent messages in bytes",
		Buckets:   c.MessageSentSizeBuckets,
	}, []string{"type"})

	MessageSentTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_sent_duration_seconds",
		Help:      "The time spent in during asynchronous message sending operation (buffer)",
		Buckets:   c.MessageSentTimeBuckets,
	}, []string{"type"})

	MessageProcessingTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_processing_duration_seconds",
		Help:      "The time spent in message handler execution",
		Buckets:   c.MessageProcessingTimeBuckets,
	}, []string{"type"})

	MessageBroadcastSentTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "message_broadcast_sent_duration_seconds",
		Help:      "The time spent in a broadcast operation",
		Buckets:   c.MessageBroadcastSentTimeBuckets,
	}, []string{"type"})

	ConcurrentHandlers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "concurrent_handlers",
		Help:      "The current number of running handlers in the server",
	}, []string{"type"})

	ConfigMaxConcurrentHandlers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goomerang",
		Subsystem: "server",
		Name:      "config_max_concurrency",
		Help:      "The configured maximum number of parallel handlers in the server",
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
	MessageInflightTimeBuckets      []float64
	MessageReceivedSizeBuckets      []float64
	MessageSentSizeBuckets          []float64
	MessageProcessingTimeBuckets    []float64
	MessageBroadcastSentTimeBuckets []float64
	MessageSentTimeBuckets          []float64
}
