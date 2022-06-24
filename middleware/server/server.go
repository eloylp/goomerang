package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/internal/messaging"
	"go.eloylp.dev/goomerang/message"
	serverMetrics "go.eloylp.dev/goomerang/metrics/server"
	"go.eloylp.dev/goomerang/middleware"
	"go.eloylp.dev/goomerang/server"
)

type MeteredServer struct {
	s       *server.Server
	metrics *serverMetrics.Metrics
}

func NewMeteredServer(m *serverMetrics.Metrics, opts ...server.Option) (*MeteredServer, error) {
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
	monitorOpts := []server.Option{
		server.WithOnStatusChangeHook(StatusMetricHook(m)),
		server.WithOnWorkerStart(WorkerStartMetricHook(m)),
		server.WithOnWorkerEnd(WorkerEndMetricHook(m)),
		server.WithOnConfiguration(ConfigurationMaxConcurrentMetricHook(m)),
		server.WithOnErrorHook(func(err error) {
			m.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...)
	s, err := server.New(mergedOpts...)
	if err != nil {
		return nil, err
	}
	s.Middleware(metricsMiddleware)
	return &MeteredServer{s: s, metrics: m}, nil
}

func (s *MeteredServer) RegisterMiddleware(m message.Middleware) {
	s.s.Middleware(m)
}

func (s *MeteredServer) RegisterHandler(msg proto.Message, handler message.Handler) {
	s.s.Handle(msg, handler)
}

func (s *MeteredServer) BroadCast(ctx context.Context, msg *message.Message) (brResult []server.BroadcastResult, err error) {
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

func StatusMetricHook(m *serverMetrics.Metrics) func(status uint32) {
	return func(status uint32) {
		m.CurrentStatus.Set(float64(status))
	}
}

func WorkerStartMetricHook(m *serverMetrics.Metrics) func() {
	return func() {
		m.ConcurrentWorkers.Inc()
	}
}

func WorkerEndMetricHook(m *serverMetrics.Metrics) func() {
	return func() {
		m.ConcurrentWorkers.Dec()
	}
}

func ConfigurationMaxConcurrentMetricHook(m *serverMetrics.Metrics) func(cfg *server.Cfg) {
	return func(cfg *server.Cfg) {
		m.ConfigMaxConcurrency.Set(float64(cfg.MaxConcurrency))
	}
}
