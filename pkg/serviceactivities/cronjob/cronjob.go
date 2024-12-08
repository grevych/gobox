package cronjob

import (
	"context"

	"github.com/grevych/gobox/pkg/async"
	"github.com/robfig/cron/v3"
)

// Job is an interface used by the Job Service Activity
// to manage cron jobs.
type Job interface {
	Run(ctx context.Context) error
}

// ServiceActivity implements the async.Runner & async.Closer interface for
// executing cron jobs.
type ServiceActivity struct {
	frequency string
	cron      *cron.Cron
	NewJob    func() Job
	done      chan struct{}
}

// Make sure ServiceActivity implements Runner interface.
var _ async.Runner = (*ServiceActivity)(nil)

// Make sure ServiceActivity implemets Closer interface.
var _ async.Closer = (*ServiceActivity)(nil)

// New creates a new job service activity runner that executes a cron job
// in a given frequency.
func New(newJob func() Job, frq string) *ServiceActivity {
	return &ServiceActivity{
		frequency: frq,
		cron:      cron.New(),
		done:      make(chan struct{}),
		NewJob:    newJob,
	}
}

// Run runs the job service activity
func (sa *ServiceActivity) Run(ctx context.Context) error {
	job := sa.NewJob()

	errCh := make(chan error, 1)
	runner := cron.FuncJob(func() {
		if err := job.Run(ctx); err != nil {
			errCh <- err
		}
	})

	if _, err := sa.cron.AddJob(sa.frequency, runner); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		sa.cron.Stop()
	}()

	sa.cron.Start()

	var err error
	select {
	// Other services either closing or failing
	case <-ctx.Done():
		err = ctx.Err()
	// Explicity closing of this service activity
	case <-sa.done:
		err = nil
	// Local error
	case err = <-errCh:
		break
	}

	ctx2 := sa.cron.Stop()

	// Wait for added jobs to finish
	<-ctx2.Done()

	// Close the job task
	err2 := async.RunClose(ctx, job)
	if err == nil {
		err = err2
	}

	return err
}

// Close closes the job service activity
func (sa *ServiceActivity) Close(_ context.Context) error {
	close(sa.done)
	return nil
}
