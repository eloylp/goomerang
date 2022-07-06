//go:build racy

package goomerang_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/kit/exec"

	"go.eloylp.dev/goomerang/client"
	"go.eloylp.dev/goomerang/internal/test"
	"go.eloylp.dev/goomerang/server"
)

// TestNoRaces tries to find nasty data races in the code.
// It works by stressing out the system on all its possible
// execution paths at the same time.
//
// IMPORTANT NOTE, this test assumes that the rest of the
// tests are already passing. Its not the target of this
// test to check functionality. Only data races. It must
// be executed in isolation.
func TestNoRaces(t *testing.T) {
	t.Parallel()

	// Bring up 2 clients and one server.
	s, run := PrepareServer(t)
	s.Handle(defaultMsg.Payload, echoHandler)
	run()

	c, connect := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c.Handle(defaultMsg.Payload, nilHandler)

	connect()

	c2, connect2 := PrepareClient(t, client.WithServerAddr(s.Addr()))
	c2.Handle(defaultMsg.Payload, nilHandler)
	connect2()

	// Set test duration of 10 seconds.
	ctx, cancl := context.WithTimeout(defaultCtx, 10*time.Second)
	defer cancl()

	wg := &sync.WaitGroup{}
	// Use the arbiter to register successes. Have at least an idea of whats happening.
	arbiter := test.NewArbiter(t)

	// Stress all parts of the system through the public API. All uses cases at the same time.
	const maxConcurrent = 20
	go exec.Parallelize(ctx, wg, maxConcurrent, func() {
		if _, err := s.BroadCast(defaultCtx, defaultMsg); err != nil && err != server.ErrNotRunning {
			arbiter.ErrorHappened(err)
			return
		}
		arbiter.ItsAFactThat("s.BroadCast()")
	})

	go exec.Parallelize(ctx, wg, maxConcurrent, func() {
		if _, err := c.Send(defaultMsg); err != nil && err != client.ErrNotRunning {
			arbiter.ErrorHappened(err)
			return
		}
		arbiter.ItsAFactThat("c.Send()")
	})

	go exec.Parallelize(ctx, wg, maxConcurrent, func() {
		if _, _, err := c.SendSync(defaultCtx, defaultMsg); err != nil && err != client.ErrNotRunning {
			arbiter.ErrorHappened(err)
			return
		}
		arbiter.ItsAFactThat("c.SendSync()")
	})

	go exec.Parallelize(ctx, wg, maxConcurrent, func() {
		if _, err := c2.Send(defaultMsg); err != nil && err != client.ErrNotRunning {
			arbiter.ErrorHappened(err)
			return
		}
		arbiter.ItsAFactThat("c2.Send()")
	})

	go exec.Parallelize(ctx, wg, maxConcurrent, func() {
		if _, _, err := c2.SendSync(defaultCtx, defaultMsg); err != nil && err != client.ErrNotRunning {
			arbiter.ErrorHappened(err)
			return
		}
		arbiter.ItsAFactThat("c2.SendSync()")
	})

	<-ctx.Done()

	// Call shutdown functions and hopefully find some data races.
	arbiter.ErrorHappened(c.Close(defaultCtx))
	arbiter.ErrorHappened(c2.Close(defaultCtx))
	arbiter.ErrorHappened(s.Shutdown(defaultCtx))

	wg.Wait()
	// Have minimum feedback of what happened.
	t.Logf("Registered errors: %v", arbiter.Errors())
	t.Logf("Registered events: %v", arbiter.EvCount())
}
