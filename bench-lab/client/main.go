package main

import (
	"context"
	"log"
	"net"
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

	"go.eloylp.dev/goomerang/client"
	metrics "go.eloylp.dev/goomerang/metrics/client"
	"go.eloylp.dev/goomerang/middleware"
)

var (
	PprofListenAddr    = os.Getenv("PPROF_LISTEN_ADDR")
	MetricsListenAddr  = os.Getenv("METRICS_LISTEN_ADDR")
	ProxyAddr          = os.Getenv("PROXY_ADDR")
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
	c, err := client.NewClient( /// todo rename this to simply New()
		client.WithTargetServer(ProxyAddr),
		client.WithMaxConcurrency(concurrency),
		client.WithOnStatusChangeHook(middleware.ClientStatusMetricHook),
		client.WithOnErrorHook(func(err error) {
			metrics.Errors.Inc()
			logrus.WithError(err).Error("internal client error detected")
		}),
	)
	mc := middleware.NewMeteredClient(c)
	if err != nil {
		log.Fatal(err)
	}
	c.RegisterHandler(&model.Point{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		logrus.Printf("client: received message : %s \n", msg.Metadata.Type)
	}))
	logrus.Infoln("starting client ...")
	mustWaitTCPService(ProxyAddr, 100*time.Millisecond, 5*time.Second)
	go func() {
		if err := mc.Connect(context.Background()); err != nil {
			logrus.WithError(err).Fatal("error connecting client")
		}
	}()

	ctx, cancl := context.WithCancel(context.Background())

	go interactions(ctx, mc)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	cancl()

	if err := mc.Close(context.Background()); err != nil {
		logrus.WithError(err).Error("error shutting down client")
	}
	logrus.Infoln("shutting down client ...")
}

func interactions(ctx context.Context, c *middleware.MeteredClient) {
	messageSize, err := strconv.Atoi(MessageSizeBytes)
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, messageSize)
	go sendMessages(ctx, c, bytes)
	go sendSyncMessages(ctx, c, bytes)
}

func sendMessages(ctx context.Context, c *middleware.MeteredClient, bytes []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := c.Send(message.New().SetPayload(&model.Point{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error sending message")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func sendSyncMessages(ctx context.Context, c *middleware.MeteredClient, bytes []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := c.SendSync(context.Background(), message.New().SetPayload(&model.Point{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error sending sync message")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func mustWaitTCPService(addr string, interval, maxWait time.Duration) {
	ctx, cancl := context.WithTimeout(context.Background(), maxWait)
	defer cancl()
	for {
		select {
		case <-ctx.Done():
			panic(ctx.Err())
		default:
			con, conErr := net.Dial("tcp", addr)
			if conErr == nil {
				_ = con.Close()
				return
			}
			time.Sleep(interval)
		}
	}
}
