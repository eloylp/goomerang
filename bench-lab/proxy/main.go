package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/toxiproxy"
	toxiClient "github.com/Shopify/toxiproxy/client"
	"github.com/sirupsen/logrus"
)

var (
	ProxyListenAddr  = os.Getenv("PROXY_LISTEN_ADDR")
	ServerListenAddr = os.Getenv("SERVER_LISTEN_ADDR")
)

func main() {
	mustWaitTCPService(ServerListenAddr, time.Millisecond, time.Second)

	proxyServer := toxiproxy.NewServer()
	go proxyServer.Listen("localhost", "8474")
	mustWaitTCPService("localhost:8474", time.Millisecond, time.Second)
	proxyClient := toxiClient.NewClient("localhost:8474")

	proxy, err := proxyClient.CreateProxy("goomerang", ProxyListenAddr, ServerListenAddr)
	if err != nil {
		logrus.Fatal(err)
	}
	_, err = proxy.AddToxic("network-latency", "latency", "upstream", 1, toxiClient.Attributes{
		"latency": 20,
	})
	if err != nil {
		logrus.Fatal(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
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
