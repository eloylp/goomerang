package middleware

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	"go.eloylp.dev/goomerang/message"
	serverMetrics "go.eloylp.dev/goomerang/metrics/server"
	"go.eloylp.dev/goomerang/server"
)

type MeteredServer struct {
	s *server.Server
}

func NewMeteredServer(s *server.Server) *MeteredServer {
	metricsMiddleware, err := PromHistograms(PromConfig{
		MessageInflightTime:   serverMetrics.MessageInflightTime,
		ReceivedMessageSize:   serverMetrics.ReceivedMessageSize,
		MessageProcessingTime: serverMetrics.MessageProcessingTime,
		SentMessageSize:       serverMetrics.SentMessageSize,
	})
	if err != nil {
		serverMetrics.Errors.Inc()
		panic(fmt.Errorf("goomerang: connect: instrumentation: %v", err))
	}
	s.RegisterMiddleware(metricsMiddleware)
	return &MeteredServer{s: s}
}

func (s *MeteredServer) RegisterMiddleware(m message.Middleware) {
	s.s.RegisterMiddleware(m)
}

func (s *MeteredServer) RegisterHandler(msg proto.Message, handler message.Handler) {
	s.s.RegisterHandler(msg, handler)
}

func (s *MeteredServer) BroadCast(ctx context.Context, msg *message.Message) (int, int, error) {
	start := time.Now()
	payloadSize, count, err := s.s.BroadCast(ctx, msg)
	if err != nil {
		serverMetrics.Errors.Inc()
		return 0, 0, err
	}
	serverMetrics.BroadcastSentTime.WithLabelValues(msg.Metadata.Type).Observe(time.Since(start).Seconds())
	for i := 0; i < count; i++ {
		serverMetrics.SentMessageSize.WithLabelValues(msg.Metadata.Type).Observe(float64(payloadSize))
	}
	return payloadSize, count, err
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
