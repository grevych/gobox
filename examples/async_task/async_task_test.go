package async_task

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTask_RunWithErrors(t *testing.T) {
	count := atomic.Int32{}
	replicas := 5
	deadline := time.Now().Add(200 * time.Millisecond)
	ctx, _ := context.WithDeadline(context.Background(), deadline)

	f := func() error {
		count.Add(1)
		return errors.New("some error")
	}
	at := NewAsyncTask(f, replicas)

	// Run with deadline since errors don't stop the task
	at.Run(ctx)
	assert.Equal(t, int32(5), count.Load())
}

func TestTask_RunWithContextCanceled(t *testing.T) {
	count := atomic.Int32{}
	replicas := 5
	ctx, cancel := context.WithCancel(context.Background())

	f := func() error {
		if count.Load() < int32(replicas) {
			count.Add(1)
		}
		return nil
	}
	at := NewAsyncTask(f, replicas)

	go func() {
		defer cancel()
		for count.Load() < int32(replicas) {
			// Make sure task ran at least once on each replica
		}
	}()

	at.Run(ctx)

	assert.Equal(t, int32(5), count.Load())
}

func TestTask_RunWithSignal(t *testing.T) {
	count := atomic.Int32{}
	replicas := 5
	wg := sync.WaitGroup{}

	f := func() error {
		if count.Load() < int32(replicas) {
			count.Add(1)
		}
		return nil
	}
	at := NewAsyncTask(f, replicas)

	wg.Add(1)
	go func() {
		defer wg.Done()
		at.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	wg.Wait()

	assert.Equal(t, int32(5), count.Load())
}
