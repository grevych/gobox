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

func TestServiceActivity_Runt(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdownSvc := New()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdownSvc.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	wg.Wait()

	assert.ErrorContains(t, shutdownErr, "process has shutdown")

	var sigErr SignalError
	assert.Assert(t, errors.As(shutdownErr, &sigErr))
	assert.Assert(t, sigErr.Signal == syscall.SIGHUP)
}

func TestServiceActivity_RuntAndCancelContext(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdownSvc := New()
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdownSvc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	wg.Wait()

	assert.ErrorContains(t, shutdownErr, "context canceled")
}

func TestServiceActivity_Close(t *testing.T) {
	var shutdownErr error
	wg := sync.WaitGroup{}
	shutdownSvc := New()
	ctx := context.Background()

	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErr = shutdownSvc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	shutdownSvc.Close(ctx)

	wg.Wait()

	assert.Assert(t, shutdownErr == nil)
}
