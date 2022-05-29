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
	c *client.Client
}

func NewMeteredClient(opts ...client.Option) (*MeteredClient, error) {
	metricsMiddleware, err := middleware.PromHistograms(middleware.PromConfig{
		MessageInflightTime:   clientMetrics.MessageInflightTime,
		MessageReceivedSize:   clientMetrics.MessageReceivedSize,
		MessageProcessingTime: clientMetrics.MessageProcessingTime,
		MessageSentSize:       clientMetrics.MessageSentSize,
		MessageSentTime:       clientMetrics.MessageSentTime,
	})
	if err != nil {
		clientMetrics.Errors.Inc()
		panic(fmt.Errorf("goomerang: connect: instrumentation: %v", err))
	}
	monitorOpts := []client.Option{
		client.WithOnStatusChangeHook(StatusMetricHook),
		client.WithOnHandlerStart(HandlerStartMetricHook),
		client.WithOnHandlerEnd(HandlerEndMetricHook),
		client.WithOnConfiguration(ConfigurationMaxConcurrentMetricHook),
		client.WithOnErrorHook(func(err error) {
			clientMetrics.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...)
	c, err := client.New(mergedOpts...)
	if err != nil {
		return nil, err
	}
	c.RegisterMiddleware(metricsMiddleware)
	return &MeteredClient{c: c}, nil
}

func (c *MeteredClient) Connect(ctx context.Context) error {
	return c.c.Connect(ctx)
}

func (c *MeteredClient) Send(msg *message.Message) (payloadSize int, err error) {
	start := time.Now()
	payloadSize, err = c.c.Send(msg)
	if err != nil {
		clientMetrics.Errors.Inc()
		return 0, err
	}
	fqdn := messaging.FQDN(msg.Payload)
	clientMetrics.MessageSentTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	clientMetrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(payloadSize))
	return
}

func (c *MeteredClient) SendSync(ctx context.Context, msg *message.Message) (payloadSize int, response *message.Message, err error) {
	start := time.Now()
	payloadSize, response, err = c.c.SendSync(ctx, msg)
	if err != nil {
		clientMetrics.Errors.Inc()
		return 0, nil, err
	}
	fqdn := messaging.FQDN(msg.Payload)
	clientMetrics.MessageSentSyncResponseTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	clientMetrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(payloadSize))
	return
}

func (c *MeteredClient) Close(ctx context.Context) (err error) {
	if err = c.c.Close(ctx); err != nil {
		clientMetrics.Errors.Inc()
	}
	return
}

func (c *MeteredClient) RegisterMiddleware(m message.Middleware) {
	c.c.RegisterMiddleware(m)
}

func (c *MeteredClient) RegisterHandler(msg proto.Message, h message.Handler) {
	c.c.RegisterHandler(msg, h)
}

func (c *MeteredClient) RegisterMessage(msg proto.Message) {
	c.c.RegisterMessage(msg)
}

func StatusMetricHook(status uint32) {
	clientMetrics.CurrentStatus.Set(float64(status))
}

func HandlerStartMetricHook(kind string) {
	clientMetrics.ConcurrentHandlers.WithLabelValues(kind).Inc()
}

func HandlerEndMetricHook(kind string) {
	clientMetrics.ConcurrentHandlers.WithLabelValues(kind).Dec()
}

func ConfigurationMaxConcurrentMetricHook(cfg *client.Cfg) {
	clientMetrics.ConfigMaxConcurrentHandlers.Set(float64(cfg.MaxConcurrency))
}
