//go:build long

package goomerang_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/kit/exec"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/example/protos"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/message"
	"go.eloylp.dev/goomerang/server"
)

func TestNoErrorsTransferringMessages(t *testing.T) {
	t.Parallel()

	// Server
	serverArbiter := test.NewArbiter(t)
	s, run := Server(t, server.WithOnErrorHook(noErrorHook(serverArbiter)))
	s.Handle(defaultMsg.Payload, message.HandlerFunc(func(sender message.Sender, msg *message.Message) {
		reply := &protos.ReplyV1{
			Message: "return back !",
		}
		if _, err := sender.Send(message.New().SetPayload(reply)); err != nil {
			serverArbiter.ErrorHappened(err)
		}
	}))
	run()
	defer s.Shutdown(defaultCtx)
	// Client
	clientArbiter := test.NewArbiter(t)

	c, connect := Client(t,
		client.WithServerAddr(s.Addr()),
		client.WithOnErrorHook(noErrorHook(clientArbiter)),
	)
	c.Handle(&protos.ReplyV1{}, nilHandler)
	connect()
	defer c.Close(defaultCtx)
	// Start sending messages
	ctx, cancel := context.WithTimeout(defaultCtx, 10*time.Second)
	defer cancel()
	wg := &sync.WaitGroup{}
	exec.Parallelize(ctx, wg, 10, func() {
		if _, err := c.Send(message.New().SetPayload(&protos.MessageV1{
			Message: "a message !",
		})); err != nil {
			serverArbiter.ErrorHappened(err)
		}
	})
	wg.Wait()
	// Ensure no errors took place in the entire execution.
	serverArbiter.RequireNoErrors()
	clientArbiter.RequireNoErrors()
}
