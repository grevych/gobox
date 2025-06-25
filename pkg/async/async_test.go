package async_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grevych/gobox/pkg/async"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

type runWithCloser struct {
	isclosed bool
}

func (r *runWithCloser) Run(c context.Context) error {
	for {
		time.Sleep(time.Millisecond * 20)
		if c.Err() != nil {
			return c.Err()
		}
	}
}
func (r *runWithCloser) Close(c context.Context) error {
	r.isclosed = true
	return nil
}

func TestRunGroupErrorPropagation(t *testing.T) {
	ctx := context.Background()
	r1 := async.Func(func(c context.Context) error {
		return fmt.Errorf("oh no")
	})
	r2 := runWithCloser{}
	aggr := async.RunGroup([]async.Runner{&r1, &r2})
	err := aggr.Run(ctx)
	assert.ErrorContains(t, err, "oh no")
	assert.Equal(t, r2.isclosed, true, "Closed the infinite loop correctly")
}

func TestRunCancelPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	async.Run(ctx, async.Func(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}))

	cancel()
	async.Default.Wait()
}

func TestRunDeadlinePropagation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	async.Run(ctx, async.Func(func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("no deadline!")
		}
		return nil
	}))

	cancel()
	async.Default.Wait()
}

func TestSleepUntil(t *testing.T) {
	now := time.Now()
	sleepTime := 50 * time.Millisecond
	async.SleepUntil(context.Background(), time.Now().Add(sleepTime))
	assert.Assert(t, time.Since(now) >= sleepTime, "slept too short")
}

func TestMutexWithContext_EarlyCancel(t *testing.T) {
	mutex := async.NewMutexWithContext()
	ctx1 := context.Background()

	err := mutex.Lock(ctx1)
	assert.NilError(t, err, "first lock acquisition should not fail")
	t.Cleanup(mutex.Unlock)

	ctx2 := context.Background()
	ctx2, cancel2 := context.WithCancel(ctx2)

	// Cancel right away, before second wait starts.
	cancel2()

	err = mutex.Lock(ctx2)
	assert.Assert(t, is.ErrorContains(err, ""), "expected error from cancellation")
}

func TestMutexWithContext_LateCancel(t *testing.T) {
	mutex := async.NewMutexWithContext()
	ctx1 := context.Background()

	err := mutex.Lock(ctx1)
	assert.NilError(t, err, "first lock acquisition should not fail")
	t.Cleanup(mutex.Unlock)

	ctx2 := context.Background()
	ctx2, cancel2 := context.WithTimeout(ctx2, time.Millisecond)
	t.Cleanup(cancel2)

	// We expect this will block until the timeout is reached.  There is
	// technically a race condition where the timeout happens before we have
	// a chance to block, in which case this test becomes equivalent to
	// `EarlyCancel`.  We can't prove that this never happens, but with a
	// long enough timeout it should be rare.
	err2 := mutex.Lock(ctx2)

	// Make sure the lock attempt was not successful.
	assert.Assert(t, is.ErrorContains(err2, ""), "expected error from cancellation")
}

func TestMutexWithContext_ExtraUnlock(t *testing.T) {
	mutex := async.NewMutexWithContext()
	ctx1 := context.Background()

	err := mutex.Lock(ctx1)
	assert.NilError(t, err, "first lock acquisition should not fail")

	mutex.Unlock()

	assert.Assert(t, is.Panics(func() { mutex.Unlock() }))
}

func TestTaskGroup_RunWithError(t *testing.T) {
	ctx := context.Background()

	counter := atomic.Int32{}
	tg := async.TaskGroup{Name: "test"}
	tg.Run(ctx, async.Func(func(ctx context.Context) error {
		counter.Add(1)
		return errors.New("some error")
	}))

	tg.Wait()
	assert.Equal(t, int32(1), counter.Load())
}

func TestTaskGroup_LoopWithError(t *testing.T) {
	ctx := context.Background()

	counter := atomic.Int32{}
	tg := async.TaskGroup{Name: "test"}
	tg.Loop(ctx, async.Func(func(ctx context.Context) error {
		if counter.Load() < 3 {
			counter.Add(1)
		} else {
			return errors.New("some error")
		}
		return nil
	}))

	tg.Wait()
	assert.Equal(t, int32(3), counter.Load())
}

func TestTaskGroup_LoopWithContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	counter := atomic.Int32{}
	tg := async.TaskGroup{Name: "test"}
	tg.Loop(ctx, async.Func(func(ctx context.Context) error {
		counter.Add(1)
		return nil
	}))

	go func() {
		for {
			if counter.Load() > 3 {
				cancel()
				break
			}
		}
	}()

	tg.Wait()
	assert.Assert(t, true)
}

func ExampleTaskGroup_run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tasks := async.TaskGroup{Name: "example"}
	tasks.Run(ctx, async.Func(func(ctx context.Context) error {
		fmt.Println("Run example")
		return nil
	}))

	cancel()
	tasks.Wait()

	// Output: Run example
}

func ExampleLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := atomic.Int32{}
	async.Loop(ctx, async.Func(func(ctx context.Context) error {
		if counter.Load() < 3 {
			counter.Add(1)
			fmt.Println("count", counter.Load())
		} else {
			<-ctx.Done()
		}
		return nil
	}))

	time.Sleep(time.Millisecond * 5)
	cancel()
	async.Default.Wait()

	// Output:
	// count 1
	// count 2
	// count 3
}

func ExampleTaskGroup_loop() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := atomic.Int32{}
	tasks := async.TaskGroup{Name: "example"}
	tasks.Loop(ctx, async.Func(func(ctx context.Context) error {
		if counter.Load() < 3 {
			counter.Add(1)
			fmt.Println("count", counter.Load())
		} else {
			<-ctx.Done()
		}
		return nil
	}))

	async.Sleep(ctx, time.Millisecond*5)
	cancel()
	tasks.Wait()

	// Output:
	// count 1
	// count 2
	// count 3
}
