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
		server.WithOnHandlerStart(HandlerStartMetricHook),
		server.WithOnHandlerEnd(HandlerEndMetricHook),
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
	s.RegisterMiddleware(metricsMiddleware)
	return &MeteredServer{s: s}, nil
}

func (s *MeteredServer) RegisterMiddleware(m message.Middleware) {
	s.s.RegisterMiddleware(m)
}

func (s *MeteredServer) RegisterHandler(msg proto.Message, handler message.Handler) {
	s.s.RegisterHandler(msg, handler)
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

func HandlerStartMetricHook(kind string) {
	serverMetrics.ConcurrentHandlers.WithLabelValues(kind).Inc()
}

func HandlerEndMetricHook(kind string) {
	serverMetrics.ConcurrentHandlers.WithLabelValues(kind).Dec()
}

func ConfigurationMaxConcurrentMetricHook(cfg *server.Cfg) {
	serverMetrics.ConfigMaxConcurrentHandlers.Set(float64(cfg.MaxConcurrency))
}
