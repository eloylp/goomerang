package goomerang_test

import (
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestUserIsCapableOfCreatingSubscriptions(t *testing.T) {
	arbiter := test.NewArbiter(t)
	s, run := PrepareServer(t, server.WithOnErrorHook(noErrorHook(arbiter)))
	groups := &sync.Map{}

	s.Handle(&protos.SubscriptionV1{}, message.HandlerFunc(func(s message.Sender, msg *message.Message) {
		subs := msg.Payload.(*protos.SubscriptionV1)

		val, ok := groups.Load(subs.SubscribeTo)
		if !ok {
			var senders []message.Sender
			senders = append(senders, s)
			groups.Store(subs.SubscribeTo, senders)
			return
		}
		senders := val.([]message.Sender)
		senders = append(senders, s)
		groups.Store(subs.SubscribeTo, senders)
	}))
	run()
	defer s.Shutdown(defaultCtx)

	c1, connect1 := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c1.Handle(&protos.MessageV1{}, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT1_RECEIVED_MESSAGE")
	}))
	connect1()
	defer c1.Close(defaultCtx)

	c2, connect2 := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c2.Handle(&protos.MessageV1{}, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT2_RECEIVED_MESSAGE")
	}))
	connect2()
	defer c2.Close(defaultCtx)

	c3, connect3 := PrepareClient(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(arbiter)),
	)
	c3.Handle(&protos.MessageV1{}, message.HandlerFunc(func(_ message.Sender, msg *message.Message) {
		_ = msg.Payload.(*protos.MessageV1)
		arbiter.ItsAFactThat("CLIENT3_RECEIVED_MESSAGE")
	}))
	connect3()
	defer c2.Close(defaultCtx)

	msg := message.New().SetPayload(&protos.SubscriptionV1{SubscribeTo: "group1"})

	_, err := c1.Send(msg)
	failIfErr(t, err)

	_, err = c2.Send(msg)
	failIfErr(t, err)

	time.Sleep(time.Second)

	senders, ok := groups.Load("group1")
	if !ok {
		t.Fatal("cannot found subscription group")
	}
	msgS := message.New().SetPayload(&protos.MessageV1{Message: "a message"})
	ss := senders.([]message.Sender)
	for i := range ss {
		_, err := ss[i].Send(msgS)
		failIfErr(t, err)
	}

	arbiter.RequireNoErrors()
	arbiter.RequireHappened("CLIENT1_RECEIVED_MESSAGE")
	arbiter.RequireHappened("CLIENT2_RECEIVED_MESSAGE")
	arbiter.RequireNotHappened("CLIENT3_RECEIVED_MESSAGE")
}
