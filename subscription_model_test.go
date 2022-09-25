//go:build integration

package goomerang_test

import (
	"testing"
	"time"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestSubscriptionsFromClientSide(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, run := Server(t, server.WithOnErrorHook(noErrorHook(arbiter)))
	s.RegisterMessage(defaultMsg().Payload)
	run()
	defer s.Shutdown(defaultCtx)
	c1, connect1 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c1.Handle(defaultMsg().Payload, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_MESSAGE")
	}))
	connect1()
	failIfErr(t, c1.Subscribe("topic.a"))
	defer c1.Close(defaultCtx)

	c2, connect2 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c2.Handle(defaultMsg().Payload, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_MESSAGE")
	}))
	connect2()
	failIfErr(t, c2.Subscribe("topic.a"))
	defer c2.Close(defaultCtx)

	c3, connect3 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	connect3()
	if _, err := c3.Publish("topic.a", defaultMsg()); err != nil {
		failIfErr(t, err)
	}
	defer c3.Close(defaultCtx)

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_MESSAGE")
	arbiter.RequireHappened("CLIENT2_RECEIVED_MESSAGE")
	arbiter.RequireNotHappened("CLIENT3_RECEIVED_MESSAGE")
}

func TestServerPublish(t *testing.T) {
	t.Parallel()

	arbiter := test.NewArbiter(t)
	s, run := Server(t, server.WithOnErrorHook(noErrorHook(arbiter)))
	run()
	defer s.Shutdown(defaultCtx)
	c1, connect1 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c1.Handle(defaultMsg().Payload, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_MESSAGE")
	}))
	connect1()
	failIfErr(t, c1.Subscribe("topic.a"))
	defer c1.Close(defaultCtx)

	c2, connect2 := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c2.Handle(defaultMsg().Payload, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_MESSAGE")
	}))
	connect2()
	failIfErr(t, c2.Subscribe("topic.a"))
	defer c2.Close(defaultCtx)

	time.Sleep(500 * time.Millisecond) // wait for subscriptions to happen

	failIfErr(t, s.Publish("topic.a", defaultMsg()))

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_MESSAGE")
	arbiter.RequireHappened("CLIENT2_RECEIVED_MESSAGE")
}
