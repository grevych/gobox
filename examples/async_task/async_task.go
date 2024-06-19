package async_task

import (
	"context"
	"time"

	"github.com/grevych/gobox/pkg/async"
)

// AsyncTask is an asynchronous task runner for functions of type Task.
type AsyncTask struct {
	*async.TaskGroup
	f        Task
	replicas int
}

// Task represents a task that needs to be always running.
type Task func() error

// NewAsyncTask creates a new instance of AsyncTask.
func NewAsyncTask(f Task, replicas int) *AsyncTask {
	taskGroup := async.NewTaskGroup("asyncTask")
	return &AsyncTask{taskGroup, f, replicas}
}

// Run executes a number of asynchronous tasks in loop and blocks
// until a shutdown signal is triggered, or the parent context is canceled. The
// number of executed tasks is defined by the replicas parameter.
func (at *AsyncTask) Run(ctx context.Context) {
	ctx2, cancel := context.WithCancel(ctx)
	shutdown := async.NewShutdown()

	go func() {
		if err := shutdown.Run(ctx2); err != nil {
			cancel()
		}
	}()

	for i := 0; i < at.replicas; i++ {
		at.Loop(ctx2, async.Func(func(ctx context.Context) error {
			if err := at.f(); err != nil {
				// We can add a recovery period
				time.Sleep(1 * time.Second)
			}

			// Return nil to keep the task running
			return nil
		}))
	}

	at.Wait()
	shutdown.Close(ctx)
}
