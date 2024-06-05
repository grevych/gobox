package async

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Shutdown is a task runner for handling graceful shutdowns.
type Shutdown struct {
	done chan struct{}
}

// Make sure Shutdown implements Runner interface.
var _ Runner = &Shutdown{}

// Make sure Shutdown implemets Closer interface.
var _ Closer = &Shutdown{}

// NewShutdown creates a new shutdown runner that listens for interrupt signals
// and handles gracefully shutting down async tasks.
func NewShutdown() *Shutdown {
	return &Shutdown{
		done: make(chan struct{}),
	}
}

// Run runs the shutdown task
func (s *Shutdown) Run(ctx context.Context) error {
	// listen for interrupt, terminated, and hangup signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case sig := <-c:
		// Allow interrupt signals to be caught again in worse-case scenario
		// situations when the service hangs during a graceful shutdown.
		signal.Reset(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		return errors.New(fmt.Sprintf("signal %v", sig))
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return nil
	}
}

// Close closes the shutdown runner
func (s *Shutdown) Close(_ context.Context) error {
	close(s.done)
	return nil
}
