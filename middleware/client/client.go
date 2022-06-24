package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	clientMetrics "go.eloylp.dev/goomerang/metrics/client"
	"go.eloylp.dev/goomerang/middleware"
)

type MeteredClient struct {
	c       *client.Client
	metrics *clientMetrics.Metrics
}

func NewMeteredClient(m *clientMetrics.Metrics, opts ...client.Option) (*MeteredClient, error) {
	metricsMiddleware, err := middleware.PromHistograms(middleware.PromConfig{
		MessageInflightTime:   m.MessageInflightTime,
		MessageReceivedSize:   m.MessageReceivedSize,
		MessageProcessingTime: m.MessageProcessingTime,
		MessageSentSize:       m.MessageSentSize,
		MessageSentTime:       m.MessageSentTime,
	})
	if err != nil {
		m.Errors.Inc()
		panic(fmt.Errorf("goomerang: connect: instrumentation: %v", err))
	}
	monitorOpts := []client.Option{
		client.WithOnStatusChangeHook(StatusMetricHook(m)),
		client.WithOnWorkerStart(WorkerStartMetricHook(m)),
		client.WithOnWorkerEnd(WorkerEndMetricHook(m)),
		client.WithOnConfiguration(ConfigurationMaxConcurrentMetricHook(m)),
		client.WithOnErrorHook(func(err error) {
			m.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...)
	c, err := client.New(mergedOpts...)
	if err != nil {
		return nil, err
	}
	c.Middleware(metricsMiddleware)
	return &MeteredClient{c: c, metrics: m}, nil
}

func (c *MeteredClient) Connect(ctx context.Context) error {
	return c.c.Connect(ctx)
}

func (c *MeteredClient) Send(msg *message.Message) (payloadSize int, err error) {
	start := time.Now()
	payloadSize, err = c.c.Send(msg)
	if err != nil {
		c.metrics.Errors.Inc()
		return 0, err
	}
	fqdn := messaging.FQDN(msg.Payload)
	c.metrics.MessageSentTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	c.metrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(payloadSize))
	return
}

func (c *MeteredClient) SendSync(ctx context.Context, msg *message.Message) (payloadSize int, response *message.Message, err error) {
	start := time.Now()
	payloadSize, response, err = c.c.SendSync(ctx, msg)
	if err != nil {
		c.metrics.Errors.Inc()
		return 0, nil, err
	}
	fqdn := messaging.FQDN(msg.Payload)
	c.metrics.MessageSentSyncResponseTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	c.metrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(payloadSize))
	return
}

func (c *MeteredClient) Close(ctx context.Context) (err error) {
	if err = c.c.Close(ctx); err != nil {
		c.metrics.Errors.Inc()
	}
	return
}

func (c *MeteredClient) RegisterMiddleware(m message.Middleware) {
	c.c.Middleware(m)
}

func (c *MeteredClient) RegisterHandler(msg proto.Message, h message.Handler) {
	c.c.Handle(msg, h)
}

func (c *MeteredClient) RegisterMessage(msg proto.Message) {
	c.c.RegisterMessage(msg)
}

func StatusMetricHook(m *clientMetrics.Metrics) func(status uint32) {
	return func(status uint32) {
		m.CurrentStatus.Set(float64(status))
	}
}

func WorkerStartMetricHook(m *clientMetrics.Metrics) func() {
	return func() {
		m.ConcurrentWorkers.Inc()
	}
}

func WorkerEndMetricHook(m *clientMetrics.Metrics) func() {
	return func() {
		m.ConcurrentWorkers.Dec()
	}
}

func ConfigurationMaxConcurrentMetricHook(m *clientMetrics.Metrics) func(cfg *client.Cfg) {
	return func(cfg *client.Cfg) {
		m.ConfigMaxConcurrency.Set(float64(cfg.MaxConcurrency))
	}
}
