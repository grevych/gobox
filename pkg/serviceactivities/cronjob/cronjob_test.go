package cronjob

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grevych/gobox/pkg/async"
	"github.com/grevych/gobox/pkg/serviceactivities/cronjobtest"
	"gotest.tools/v3/assert"
)

func TestServiceActivity_RunAndClose(t *testing.T) {
	/*

		Explanation:

			       0s                  1s                  2s                 3s
			       |-------------------|-------------------|------------------|

		Job
		Schedule
			                           < ~500 ms >
			                           ^
			                           First job exec

			                                                < ~500 ms >
			                                                ^
			                                                Second job exec

		Test
		Execution
			             <------------   ~2000 ms   ------------>
			             ^
			             Test sleep
			                                                     .......
			                                                     ^     ^
			                                                     |     |
			                                                     |     |
			                                                     |
			                                                     |     End of second job exec
			                                                     |

			                                                     Close service activity
	*/
	var cronjobErr error
	wg := sync.WaitGroup{}
	counter := atomic.Int32{}
	newJob := func() Job {
		var job async.Func = func(ctx context.Context) error {
			counter.Add(1)
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

	time.Sleep(2000 * time.Millisecond)
	cronjobSvc.Close(context.Background())

	wg.Wait()

	// Since the job was scheduled twice, we strictly wait for the
	// last job task execution.
	assert.Equal(t, counter.Load(), int32(2))
	assert.NilError(t, cronjobErr)
}

func TestServiceActivity_RuntAndCancelContext(t *testing.T) {
	/*

		Explanation:

			       0s                  1s                  2s                 3s
		           |-------------------|-------------------|------------------|

		Job
		Schedule
			                           < ~500 ms >
			                           ^
			                           First job exec

			                                                < ~500 ms >
			                                                ^
			                                                Second job exec

		Test
		Execution
			             <------------   ~2000 ms   ------------>
			             ^
			             Test sleep
			                                                     .......
			                                                     ^     ^
			                                                     |     |
			                                                     |     |
			                                                     |
			                                                     |     End of second job exec
			                                                     |

			                                                     Context canceled
	*/
	var cronjobErr error
	wg := sync.WaitGroup{}
	counter := atomic.Int32{}
	newJob := func() Job {
		var job async.Func = func(ctx context.Context) error {
			counter.Add(1)
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

	time.Sleep(2000 * time.Millisecond)
	cancel()

	wg.Wait()

	// Since the job was scheduled twice, we strictly wait for the
	// last job task execution.
	assert.Equal(t, counter.Load(), int32(2))
	assert.ErrorContains(t, cronjobErr, "context canceled")
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

func TestServiceActivity_RunAndCloseWithErrorOnClose(t *testing.T) {
	var cronjobErr error
	wg := sync.WaitGroup{}
	newJob := func() Job {
		return &cronjobtest.RunnerWithCloseError{}
	}
	cronjobSvc := New(newJob, "@every 1s")

	wg.Add(1)
	go func() {
		defer wg.Done()
		cronjobErr = cronjobSvc.Run(context.Background())
	}()

	time.Sleep(1000 * time.Millisecond)
	cronjobSvc.Close(context.Background())

	wg.Wait()

	assert.ErrorContains(t, cronjobErr, "error while closing runner")
}

func TestServiceActivity_RunAndCloseWithErrors(t *testing.T) {
	var cronjobErr error
	wg := sync.WaitGroup{}
	newJob := func() Job {
		return &cronjobtest.RunnerWithErrors{}
	}
	cronjobSvc := New(newJob, "@every 1s")

	wg.Add(1)
	go func() {
		defer wg.Done()
		cronjobErr = cronjobSvc.Run(context.Background())
	}()

	time.Sleep(1000 * time.Millisecond)
	cronjobSvc.Close(context.Background())

	wg.Wait()

	assert.ErrorContains(t, cronjobErr, "error while running runner")
}
