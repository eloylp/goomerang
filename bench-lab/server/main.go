package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.eloylp.dev/goomerang/bench-lab/model"
	"go.eloylp.dev/goomerang/message"
	"google.golang.org/protobuf/types/known/timestamppb"

	metrics "go.eloylp.dev/goomerang/metrics/server"
	"go.eloylp.dev/goomerang/middleware"
	"go.eloylp.dev/goomerang/server"
)

var (
	ListenAddr        = os.Getenv("LISTEN_ADDR")
	PprofListenAddr   = os.Getenv("PPROF_LISTEN_ADDR")
	MetricsListenAddr = os.Getenv("METRICS_LISTEN_ADDR")
	MessageSizeBytes  = os.Getenv("MESSAGE_SIZE_BYTES")
)

func main() {
	go func() {
		logrus.Println(http.ListenAndServe(PprofListenAddr, nil))
	}()
	go func() {
		logrus.Println(http.ListenAndServe(MetricsListenAddr, promhttp.Handler()))
	}()
	s, err := server.NewServer(
		server.WithListenAddr(ListenAddr),
		server.WithMaxConcurrency(10),
		server.WithOnStatusChangeHook(middleware.ServerStatusMetricHook),
		server.WithOnErrorHook(func(err error) {
			metrics.Errors.Inc()
			logrus.WithError(err).Error("internal server error detected")
		}),
	)
	ms := middleware.NewMeteredServer(s)
	if err != nil {
		log.Fatal(err)
	}
	logrus.Infoln("starting server ...")

	go func() {
		if err := ms.Run(); err != nil {
			logrus.WithError(err).Fatal("error running server")
		}
	}()
	ctx, cancl := context.WithCancel(context.Background())
	go interactions(ctx, ms)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	cancl()

	if err := ms.Shutdown(context.Background()); err != nil {
		logrus.WithError(err).Error("error shutting down server")
	}
	logrus.Infoln("shutting down server ...")
}

func interactions(ctx context.Context, s *middleware.MeteredServer) {
	messageSize, err := strconv.Atoi(MessageSizeBytes)
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, messageSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := s.BroadCast(context.Background(), message.New().SetPayload(&model.Point{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error broadcasting message")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
