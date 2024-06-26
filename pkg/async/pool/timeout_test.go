package pool_test

import (
	"context"
	"testing"
	"time"

	"github.com/grevych/gobox/pkg/async"
	"github.com/grevych/gobox/pkg/async/pool"
	"gotest.tools/v3/assert"
)

func TestTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p := pool.New(ctx, pool.ConstantSize(1))

	timeout := 5 * time.Millisecond
	scheduler := pool.WithTimeout(timeout, p)

	scheduler, wait := pool.WithWait(scheduler)

	var err error
	// The signal to start first scheduled task.
	startFirst := make(chan bool)
	defer close(startFirst)

	scheduler.Schedule(ctx, async.Func(func(ctx context.Context) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		<-startFirst
		// Ensure the next task times out.
		time.Sleep(2 * timeout)
		return nil
	}))
	scheduler.Schedule(ctx, async.Func(func(ctx context.Context) error {
		if ctx.Err() != nil {
			// Capture context error in enclosing scope.
			err = ctx.Err()
			return ctx.Err()
		}
		return nil
	}))
	startFirst <- true

	wait()

	assert.Equal(t, context.DeadlineExceeded, err)
}
