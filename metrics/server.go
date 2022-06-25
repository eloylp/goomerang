package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type ServerConfig struct {
	MessageInflightTimeBuckets      []float64
	MessageReceivedSizeBuckets      []float64
	MessageSentSizeBuckets          []float64
	MessageProcessingTimeBuckets    []float64
	MessageBroadcastSentTimeBuckets []float64
	MessageSentTimeBuckets          []float64
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		MessageInflightTimeBuckets:      prometheus.DefBuckets,
		MessageReceivedSizeBuckets:      defBucketsForSizeBytes,
		MessageSentSizeBuckets:          defBucketsForSizeBytes,
		MessageSentTimeBuckets:          prometheus.DefBuckets,
		MessageProcessingTimeBuckets:    prometheus.DefBuckets,
		MessageBroadcastSentTimeBuckets: prometheus.DefBuckets,
	}
}

type ServerMetrics struct {
	MessageInflightTime      *prometheus.HistogramVec
	MessageReceivedSize      *prometheus.HistogramVec
	MessageSentSize          *prometheus.HistogramVec
	MessageSentTime          *prometheus.HistogramVec
	MessageProcessingTime    *prometheus.HistogramVec
	MessageBroadcastSentTime *prometheus.HistogramVec
	CurrentStatus            prometheus.Gauge
	ConcurrentWorkers        prometheus.Gauge
	ConfigMaxConcurrency     prometheus.Gauge
	Errors                   prometheus.Counter
}

func NewServerMetrics(c ServerConfig) *ServerMetrics {
	return &ServerMetrics{

		MessageInflightTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_received_inflight_duration_seconds",
			Help:      "The time the message spent over the wire till received",
			Buckets:   c.MessageInflightTimeBuckets,
		}, []string{"kind"}),

		MessageReceivedSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_received_size_bytes",
			Help:      "The size of the received messages in bytes",
			Buckets:   c.MessageReceivedSizeBuckets,
		}, []string{"kind"}),

		MessageSentSize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_sent_size_bytes",
			Help:      "The size of the sent messages in bytes",
			Buckets:   c.MessageSentSizeBuckets,
		}, []string{"kind"}),

		MessageSentTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_sent_duration_seconds",
			Help:      "The time spent in during asynchronous message sending operation (buffer)",
			Buckets:   c.MessageSentTimeBuckets,
		}, []string{"kind"}),

		MessageProcessingTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_processing_duration_seconds",
			Help:      "The time spent in message handler execution",
			Buckets:   c.MessageProcessingTimeBuckets,
		}, []string{"kind"}),

		MessageBroadcastSentTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "message_broadcast_sent_duration_seconds",
			Help:      "The time spent in a broadcast operation",
			Buckets:   c.MessageBroadcastSentTimeBuckets,
		}, []string{"kind"}),

		ConcurrentWorkers: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "concurrent_workers",
			Help:      "The current number of running workers in the server",
		}),

		ConfigMaxConcurrency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "config_max_concurrency",
			Help:      "The configured maximum number of parallel workers in the server",
		}),

		CurrentStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "status",
			Help:      "The current status of the server (0 => New, 1 => Running, 2=> Closing, 3 => closed)",
		}),

		Errors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "goomerang",
			Subsystem: "server",
			Name:      "errors_total",
			Help:      "The errors happened in server",
		}),
	}
}

func (m *ServerMetrics) Register(r prometheus.Registerer) {
	r.MustRegister(m.MessageInflightTime)
	r.MustRegister(m.MessageReceivedSize)
	r.MustRegister(m.MessageSentSize)
	r.MustRegister(m.MessageSentTime)
	r.MustRegister(m.MessageProcessingTime)
	r.MustRegister(m.MessageBroadcastSentTime)
	r.MustRegister(m.ConcurrentWorkers)
	r.MustRegister(m.ConfigMaxConcurrency)
	r.MustRegister(m.CurrentStatus)
	r.MustRegister(m.Errors)
}
