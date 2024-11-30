package cronjob

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/grevych/gobox/pkg/async"
	"gotest.tools/v3/assert"
)

func TestServiceActivity_RunAndClose(t *testing.T) {
	var cronjobErr error
	wg := sync.WaitGroup{}
	counter := 0
	newJob := func() Job {
		var job async.Func = func(ctx context.Context) error {
			counter += 1
			time.Sleep(500 * time.Millisecond)
			return nil
		}
		return job
	}
	cronjobSvc := New(newJob, "@every 1s")

	wg.Add(1)
	go func() {
		defer wg.Done()
		cronjobErr = cronjobSvc.Run(context.Background())
	}()

	time.Sleep(2500 * time.Millisecond)
	cronjobSvc.Close(context.Background())

	wg.Wait()

	// Job was executed twice since we force to wait for the
	// last job schedule.
	assert.Assert(t, counter == 2)
	assert.NilError(t, cronjobErr)
}

func TestServiceActivity_RunWithError(t *testing.T) {
	var cronjobErr error
	wg := sync.WaitGroup{}
	newJob := func() Job {
		var job async.Func = func(ctx context.Context) error {
			return errors.New("cronjob with error")
		}
		return job
	}
	cronjobSvc := New(newJob, "@every 1s")

	wg.Add(1)
	go func() {
		defer wg.Done()
		cronjobErr = cronjobSvc.Run(context.Background())
	}()

	wg.Wait()

	assert.ErrorContains(t, cronjobErr, "cronjob with error")
}

func TestServiceActivity_RuntAndCancelContext(t *testing.T) {
	var cronjobErr error
	wg := sync.WaitGroup{}
	counter := 0
	newJob := func() Job {
		var job async.Func = func(ctx context.Context) error {
			counter += 1
			time.Sleep(500 * time.Millisecond)
			return nil
		}
		return job
	}
	cronjobSvc := New(newJob, "@every 1s")
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		cronjobErr = cronjobSvc.Run(ctx)
	}()

	time.Sleep(2500 * time.Millisecond)
	cancel()

	wg.Wait()

	// Job was executed twice since we force to wait for the
	// last job schedule.
	assert.Assert(t, counter == 2)
	assert.ErrorContains(t, cronjobErr, "context canceled")
}
