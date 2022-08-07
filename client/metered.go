package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/metrics"
	"go.eloylp.dev/goomerang/middleware"
)

type MeteredClient struct {
	c       *Client
	metrics *metrics.ClientMetrics
}

func NewMetered(m *metrics.ClientMetrics, opts ...Option) (*MeteredClient, error) {
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
	monitorOpts := []Option{
		WithOnStatusChangeHook(statusMetricHook(m)),
		WithOnWorkerStart(workerStartMetricHook(m)),
		WithOnWorkerEnd(workerEndMetricHook(m)),
		WithOnConfiguration(configurationMaxConcurrentMetricHook(m)),
		WithOnErrorHook(func(err error) {
			m.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...) //nolint: gocritic
	c, err := New(mergedOpts...)
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

func (c *MeteredClient) Broadcast(msg *message.Message) error {
	if _, err := c.c.Broadcast(msg); err != nil {
		c.metrics.Errors.Inc()
		return err
	}
	fqdn := messaging.FQDN(msg.Payload)
	c.metrics.BroadcastCmdCount.WithLabelValues(fqdn).Inc()
	return nil
}

func (c *MeteredClient) Subscribe(topic string) error {
	if err := c.c.Subscribe(topic); err != nil {
		c.metrics.Errors.Inc()
		return err
	}
	c.metrics.SubscribeCmdCount.WithLabelValues(topic).Inc()
	return nil
}

func (c *MeteredClient) Publish(topic string, msg *message.Message) error {
	if _, err := c.c.Publish(topic, msg); err != nil {
		c.metrics.Errors.Inc()
		return err
	}
	fqdn := messaging.FQDN(msg.Payload)
	c.metrics.PublishCmdCount.WithLabelValues(topic, fqdn).Inc()
	return nil
}

func (c *MeteredClient) Unsubscribe(topic string) error {
	if err := c.c.Unsubscribe(topic); err != nil {
		c.metrics.Errors.Inc()
		return err
	}
	c.metrics.UnsubscribeCmdCount.WithLabelValues(topic).Inc()
	return nil
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

func (c *MeteredClient) Middleware(m message.Middleware) {
	c.c.Middleware(m)
}

func (c *MeteredClient) Handle(msg proto.Message, h message.Handler) {
	c.c.Handle(msg, h)
}

func (c *MeteredClient) RegisterMessage(msg proto.Message) {
	c.c.RegisterMessage(msg)
}

func statusMetricHook(m *metrics.ClientMetrics) func(status uint32) {
	return func(status uint32) {
		m.CurrentStatus.Set(float64(status))
	}
}

func workerStartMetricHook(m *metrics.ClientMetrics) func() {
	return func() {
		m.ConcurrentWorkers.Inc()
	}
}

func workerEndMetricHook(m *metrics.ClientMetrics) func() {
	return func() {
		m.ConcurrentWorkers.Dec()
	}
}

func configurationMaxConcurrentMetricHook(m *metrics.ClientMetrics) func(cfg *Cfg) {
	return func(cfg *Cfg) {
		m.ConfigMaxConcurrency.Set(float64(cfg.MaxConcurrency))
	}
}
