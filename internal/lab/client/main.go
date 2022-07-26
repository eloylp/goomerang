package main

import (
	"context"
	"log"
	"math/rand"
	"net"
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
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.eloylp.dev/goomerang/client"
)

var (
	PprofListenAddr    = os.Getenv("PPROF_LISTEN_ADDR")
	MetricsListenAddr  = os.Getenv("METRICS_LISTEN_ADDR")
	TargetAddr         = os.Getenv("TARGET_ADDR")
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

	met := metrics.NewClientMetrics(metrics.DefaultClientConfig())
	met.Register(prometheus.DefaultRegisterer)

	mc, err := client.NewMetered(
		met,
		client.WithServerAddr(TargetAddr),
		client.WithMaxConcurrency(concurrency),
		client.WithOnErrorHook(func(err error) {
			logrus.WithError(err).Error("internal client error detected")
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	mc.Handle(&protos.PointV1{}, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		// It's ok, discard from buffer
		time.Sleep(20 * time.Millisecond)
	}))
	mc.Handle(&protos.PointReplyV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		// It's ok, discard from buffer
		time.Sleep(20 * time.Millisecond)
	}))
	mc.Handle(&protos.BroadcastV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		// process server broadcast
		time.Sleep(20 * time.Millisecond)
	}))
	logrus.Infoln("starting client ...")
	mustWaitTCPService(TargetAddr, 100*time.Millisecond, 5*time.Second)
	if err := mc.Connect(context.Background()); err != nil {
		logrus.WithError(err).Fatal("error connecting client")
	}

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

func interactions(ctx context.Context, c *client.MeteredClient) {
	messageSize, err := strconv.Atoi(MessageSizeBytes)
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, messageSize)
	if err := c.Subscribe("topic.a"); err != nil {
		panic(err)
	}
	if err := c.Subscribe("topic.b"); err != nil {
		panic(err)
	}
	go publishMessages(ctx, c, bytes)
	go sendMessages(ctx, c, bytes)
	go sendSyncMessages(ctx, c, bytes)
}

func publishMessages(ctx context.Context, c *client.MeteredClient, bytes []byte) {
	topics := []string{"topic.a", "topic.b"}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			topic := topics[rand.Intn(2)]
			err := c.Publish(topic, message.New().SetPayload(&protos.PointV1{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error sending message")
			}
		}
	}
}

func sendMessages(ctx context.Context, c *client.MeteredClient, bytes []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := c.Send(message.New().SetPayload(&protos.PointV1{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error sending message")
			}
		}
	}
}

func sendSyncMessages(ctx context.Context, c *client.MeteredClient, bytes []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := c.SendSync(context.Background(), message.New().SetPayload(&protos.PointV1{
				X:          34.45,
				Y:          89.12,
				Time:       timestamppb.Now(),
				DeviceData: bytes,
			}))
			if err != nil {
				logrus.WithError(err).Error("error sending sync message")
			}
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
