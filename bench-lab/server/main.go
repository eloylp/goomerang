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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.eloylp.dev/goomerang/bench-lab/model"
	"go.eloylp.dev/goomerang/message"
	serverMetrics "go.eloylp.dev/goomerang/metrics/server"
	serverMiddleware "go.eloylp.dev/goomerang/middleware/server"
	"go.eloylp.dev/goomerang/server"
)

var (
	ListenAddr         = os.Getenv("LISTEN_ADDR")
	PprofListenAddr    = os.Getenv("PPROF_LISTEN_ADDR")
	MetricsListenAddr  = os.Getenv("METRICS_LISTEN_ADDR")
	MessageSizeBytes   = os.Getenv("MESSAGE_SIZE_BYTES")
	HandlerConcurrency = os.Getenv("HANDLER_CONCURRENCY")
)

func main() {
	go func() {
		logrus.Println(http.ListenAndServe(PprofListenAddr, nil))
	}()
	go func() {
		logrus.Println(http.ListenAndServe(MetricsListenAddr, promhttp.Handler()))
	}()
	concurrency, err := strconv.Atoi(HandlerConcurrency)
	if err != nil {
		panic(err)
	}
	metrics := serverMetrics.NewMetrics(serverMetrics.DefaultConfig())
	metrics.Register(prometheus.DefaultRegisterer)
	ms, err := serverMiddleware.NewMeteredServer(
		metrics,
		server.WithListenAddr(ListenAddr),
		server.WithMaxConcurrency(concurrency),
		server.WithOnErrorHook(func(err error) {
			logrus.WithError(err).Error("internal server error detected")
		}),
	)
	ms.RegisterHandler(&model.PointV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		time.Sleep(20 * time.Millisecond)
		reply := message.New().SetPayload(&model.PointReplyV1{Status: "OK"})
		_, err := s.Send(reply)
		if err != nil {
			log.Fatal(err)
		}
		//logrus.Printf("server: received message : %s \n", msg.metadata.Kind)
	}))
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

func interactions(ctx context.Context, s *serverMiddleware.MeteredServer) {
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
			_, err := s.BroadCast(context.Background(), message.New().SetPayload(&model.BroadcastV1{
				Message: "Broadcasting !",
				Data:    bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error broadcasting message")
			}
		}
	}
}
