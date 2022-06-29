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
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/metrics"
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
	met := metrics.NewServerMetrics(metrics.DefaultServerConfig())
	met.Register(prometheus.DefaultRegisterer)
	ms, err := server.NewMetered(
		met,
		server.WithListenAddr(ListenAddr),
		server.WithMaxConcurrency(concurrency),
		server.WithOnErrorHook(func(err error) {
			logrus.WithError(err).Error("internal server error detected")
		}),
	)
	ms.RegisterHandler(&protos.PointV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		time.Sleep(20 * time.Millisecond)
		reply := message.New().SetPayload(&protos.PointReplyV1{Status: "OK"})
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

func interactions(ctx context.Context, s *server.MeteredServer) {
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
			_, err := s.BroadCast(context.Background(), message.New().SetPayload(&protos.BroadcastV1{
				Message: "Broadcasting !",
				Data:    bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error broadcasting message")
			}
		}
	}
}
