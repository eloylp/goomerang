package server

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

type MeteredServer struct {
	s       *Server
	metrics *metrics.ServerMetrics
}

func NewMetered(m *metrics.ServerMetrics, opts ...Option) (*MeteredServer, error) {
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
		WithOnSubscribeHook(subscribeHook(m)),
		WithOnUnsubscribeHook(unsubscribeHook(m)),
		WithOnPublishHook(publishHook(m)),
		WithOnErrorHook(func(err error) {
			m.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...) //nolint: gocritic
	s, err := New(mergedOpts...)
	if err != nil {
		return nil, err
	}
	s.Middleware(metricsMiddleware)
	return &MeteredServer{s: s, metrics: m}, nil
}

func (s *MeteredServer) Middleware(m message.Middleware) {
	s.s.Middleware(m)
}

func (s *MeteredServer) Handle(msg proto.Message, handler message.Handler) {
	s.s.Handle(msg, handler)
}

func (s *MeteredServer) Publish(topic string, msg *message.Message) error {
	if err := s.s.Publish(topic, msg); err != nil {
		s.metrics.Errors.Inc()
		return err
	}
	s.metrics.PublishCount.WithLabelValues(topic, messaging.FQDN(msg.Payload)).Inc()
	return nil
}

func (s *MeteredServer) BroadCast(ctx context.Context, msg *message.Message) (brResult []BroadcastResult, err error) {
	start := time.Now()
	brResult, err = s.s.BroadCast(ctx, msg)
	if err != nil {
		s.metrics.Errors.Inc()
		return
	}
	fqdn := messaging.FQDN(msg.Payload)
	s.metrics.MessageBroadcastSentTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	for i := 0; i < len(brResult); i++ {
		s.metrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(brResult[i].Size))
		s.metrics.MessageSentTime.WithLabelValues(fqdn).Observe(brResult[i].Duration.Seconds())
	}
	return
}

func (s *MeteredServer) Run() (err error) {
	if err = s.s.Run(); err != nil {
		s.metrics.Errors.Inc()
	}
	return
}

func (s *MeteredServer) Shutdown(ctx context.Context) (err error) {
	if err = s.s.Shutdown(ctx); err != nil {
		s.metrics.Errors.Inc()
	}
	return
}

func statusMetricHook(m *metrics.ServerMetrics) func(status uint32) {
	return func(status uint32) {
		m.CurrentStatus.Set(float64(status))
	}
}

func workerStartMetricHook(m *metrics.ServerMetrics) func() {
	return func() {
		m.ConcurrentWorkers.Inc()
	}
}

func workerEndMetricHook(m *metrics.ServerMetrics) func() {
	return func() {
		m.ConcurrentWorkers.Dec()
	}
}

func configurationMaxConcurrentMetricHook(m *metrics.ServerMetrics) func(cfg *Cfg) {
	return func(cfg *Cfg) {
		m.ConfigMaxConcurrency.Set(float64(cfg.MaxConcurrency))
	}
}

func subscribeHook(m *metrics.ServerMetrics) func(topic string) {
	return func(topic string) {
		m.SubscribeCount.WithLabelValues(topic).Inc()
	}
}

func unsubscribeHook(m *metrics.ServerMetrics) func(topic string) {
	return func(topic string) {
		m.UnsubscribeCount.WithLabelValues(topic).Inc()
	}
}

func publishHook(m *metrics.ServerMetrics) func(topic, fqdn string) {
	return func(topic, fqdn string) {
		m.PublishCount.WithLabelValues(topic, fqdn).Inc()
	}
}
