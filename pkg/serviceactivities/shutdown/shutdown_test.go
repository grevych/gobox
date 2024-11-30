package shutdown

import (
	"context"
	"errors"
	"sync"
	"syscall"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestServiceActivity_RuntWithContextDone(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := New()
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

func TestServiceActivity_RuntWithSignal(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := New()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdown.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	wg.Wait()

	assert.ErrorContains(t, shutdownErr, "process has shutdown")

	var sigErr SignalError
	assert.Assert(t, errors.As(shutdownErr, &sigErr))
	assert.Assert(t, sigErr.Signal == syscall.SIGHUP)
}

func TestServiceActivity_RuntWithClose(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdown := New()
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
