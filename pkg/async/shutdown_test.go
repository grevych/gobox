package async_test

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/grevych/gobox/pkg/async"
	"gotest.tools/v3/assert"
)

func TestShutdown_RuntWithContextDone(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := async.NewShutdown()
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdown.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	wg.Wait()

	assert.ErrorContains(t, shutdownErr, "context canceled")
}

func TestShutdown_RuntWithSignal(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := async.NewShutdown()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdown.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	wg.Wait()

	assert.ErrorContains(t, shutdownErr, "signal hangup")
}

func TestShutdown_RuntWithClose(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := async.NewShutdown()
	ctx := context.Background()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdown.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	shutdown.Close(ctx)

	wg.Wait()

	assert.Assert(t, shutdownErr == nil)
}
