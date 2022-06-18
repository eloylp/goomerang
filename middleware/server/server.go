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
	s *server.Server
}

func NewMeteredServer(opts ...server.Option) (*MeteredServer, error) {
	metricsMiddleware, err := middleware.PromHistograms(middleware.PromConfig{
		MessageInflightTime:   serverMetrics.MessageInflightTime,
		MessageReceivedSize:   serverMetrics.MessageReceivedSize,
		MessageProcessingTime: serverMetrics.MessageProcessingTime,
		MessageSentSize:       serverMetrics.MessageSentSize,
		MessageSentTime:       serverMetrics.MessageSentTime,
	})
	if err != nil {
		serverMetrics.Errors.Inc()
		panic(fmt.Errorf("goomerang: connect: instrumentation: %v", err))
	}
	monitorOpts := []server.Option{
		server.WithOnStatusChangeHook(StatusMetricHook),
		server.WithOnWorkerStart(WorkerStartMetricHook),
		server.WithOnWorkerEnd(WorkerEndMetricHook),
		server.WithOnConfiguration(ConfigurationMaxConcurrentMetricHook),
		server.WithOnErrorHook(func(err error) {
			serverMetrics.Errors.Inc()
		}),
	}
	mergedOpts := append(monitorOpts, opts...)
	s, err := server.New(mergedOpts...)
	if err != nil {
		return nil, err
	}
	s.Middleware(metricsMiddleware)
	return &MeteredServer{s: s}, nil
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
		serverMetrics.Errors.Inc()
		return
	}
	fqdn := messaging.FQDN(msg.Payload)
	serverMetrics.MessageBroadcastSentTime.WithLabelValues(fqdn).Observe(time.Since(start).Seconds())
	for i := 0; i < len(brResult); i++ {
		serverMetrics.MessageSentSize.WithLabelValues(fqdn).Observe(float64(brResult[i].Size))
		serverMetrics.MessageSentTime.WithLabelValues(fqdn).Observe(brResult[i].Duration.Seconds())
	}
	return
}

func (s *MeteredServer) Run() (err error) {
	if err = s.s.Run(); err != nil {
		serverMetrics.Errors.Inc()
	}
	return
}

func (s *MeteredServer) Shutdown(ctx context.Context) (err error) {
	if err = s.s.Shutdown(ctx); err != nil {
		serverMetrics.Errors.Inc()
	}
	return
}

func StatusMetricHook(status uint32) {
	serverMetrics.CurrentStatus.Set(float64(status))
}

func WorkerStartMetricHook() {
	serverMetrics.ConcurrentWorkers.Inc()
}

func WorkerEndMetricHook() {
	serverMetrics.ConcurrentWorkers.Dec()
}

func ConfigurationMaxConcurrentMetricHook(cfg *server.Cfg) {
	serverMetrics.ConfigMaxConcurrency.Set(float64(cfg.MaxConcurrency))
}
