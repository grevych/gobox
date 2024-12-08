package cronjobtest

import (
	"context"
	"errors"
	"time"
)

type RunnerWithCloseError struct {
}

func (r *RunnerWithCloseError) Run(ctx context.Context) error {
	time.Sleep(300 * time.Millisecond)
	return nil
}

func (r *RunnerWithCloseError) Close(ctx context.Context) error {
	return errors.New("error while closing runner")
}

type RunnerWithErrors struct {
}

func (r *RunnerWithErrors) Run(ctx context.Context) error {
	return errors.New("error while running runner")
}

func (r *RunnerWithErrors) Close(ctx context.Context) error {
	return errors.New("error while closing runner")
}
