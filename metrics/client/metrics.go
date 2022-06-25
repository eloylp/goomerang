package client

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defBucketsForSizeBytes = []float64{10, 20, 40, 80, 160, 320, 640, 1280, 2560, 5120, 10240, 50000}
)

type Config struct {
	MessageInflightTimeBuckets   []float64
	ReceivedMessageSizeBuckets   []float64
	SentMessageSizeBuckets       []float64
	MessageProcessingTimeBuckets []float64
	SendSyncResponseTimeBuckets  []float64
	SendTimeBuckets              []float64
}

func DefaultConfig() Config {
	return Config{
		MessageInflightTimeBuckets:   prometheus.DefBuckets,
		ReceivedMessageSizeBuckets:   defBucketsForSizeBytes,
		SentMessageSizeBuckets:       defBucketsForSizeBytes,
		MessageProcessingTimeBuckets: prometheus.DefBuckets,
		SendSyncResponseTimeBuckets:  prometheus.DefBuckets,
		SendTimeBuckets:              prometheus.DefBuckets,
	}
}

type Metrics struct {
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
}

func NewMetrics(c Config) *Metrics {
	return &Metrics{

		MessageInflightTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_received_inflight_duration_seconds",
			Help:      "The time the message spent over the wire till received",
			Buckets:   c.MessageInflightTimeBuckets,
		}, []string{"kind"}),

		MessageReceivedSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_received_size_bytes",
			Help:      "The size of the received messages in bytes",
			Buckets:   c.ReceivedMessageSizeBuckets,
		}, []string{"kind"}),

		MessageSentSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_sent_size_bytes",
			Help:      "The size of the sent messages in bytes",
			Buckets:   c.SentMessageSizeBuckets,
		}, []string{"kind"}),

		MessageSentTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_sent_duration_seconds",
			Help:      "The time spent in during asynchronous message sending operation (buffer)",
			Buckets:   c.SendTimeBuckets,
		}, []string{"kind"}),

		MessageProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_processing_duration_seconds",
			Help:      "The time spent in message handler execution",
			Buckets:   c.MessageProcessingTimeBuckets,
		}, []string{"kind"}),

		MessageSentSyncResponseTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "message_sent_sync_duration_seconds",
			Help:      "The time spent in a synchronous message sending operation",
			Buckets:   c.SendSyncResponseTimeBuckets,
		}, []string{"kind"}),

		CurrentStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "status",
			Help:      "The current status of the client (0 => New, 1 => Running, 2=> Closing, 3 => closed)",
		}),

		ConcurrentWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "concurrent_workers",
			Help:      "The current number of running workers in the client",
		}),

		ConfigMaxConcurrency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "config_max_concurrency",
			Help:      "The configured maximum number of parallel workers in the client",
		}),

		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "goomerang",
			Subsystem: "client",
			Name:      "errors_total",
			Help:      "The errors happened in client",
		}),
	}
}

func (m *Metrics) Register(r prometheus.Registerer) {
	r.MustRegister(m.MessageInflightTime)
	r.MustRegister(m.MessageReceivedSize)
	r.MustRegister(m.MessageSentSize)
	r.MustRegister(m.MessageSentTime)
	r.MustRegister(m.MessageProcessingTime)
	r.MustRegister(m.MessageSentSyncResponseTime)
	r.MustRegister(m.CurrentStatus)
	r.MustRegister(m.ConcurrentWorkers)
	r.MustRegister(m.ConfigMaxConcurrency)
	r.MustRegister(m.Errors)
}
