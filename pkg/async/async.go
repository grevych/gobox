// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Provides utilities for writing async code

// Package async has helper utilities for running async code with
// proper tracing.
//
// When starting go routines, use async.Run:
//
//	// this starts a go routine
//	async.Run(ctx, async.Func(func(ctx context.Context) error {
//	   ctx = trace.StartCall(ctx, "...")
//	   defer trace.EndCall(ctx)
//	   .... do whatever needs to be done ...
//	   .... just return error to abort loop ....
//	}))
//
// If the long running activity involves fetching from a reader or
// some other iterative pattern, use async.Loop
//
//	// this starts a go routine
//	async.Loop(ctx, async.Func(func(ctx context.Context) error {
//	    ctx = trace.StartCall(ctx, "...")
//	    defer trace.EndCall(ctx)
//	   .... do whatever... return errors to abort loop
//	   ... maybe async.Sleep(ctx, time.Minute)
//	}))
//
// Note:  async.Loop terminates the loop when the inner function
// returns an error or the context is canceled.
//
// To wait for all go-routines to terminate, use Tasks.Wait.  See
// examples for using Tasks
package async

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/grevych/gobox/pkg/events"
	"github.com/grevych/gobox/pkg/log"
	"github.com/grevych/gobox/pkg/trace"
)

// Runner is the default interface for a runner function
type Runner interface {
	Run(ctx context.Context) error
}

// Closer is the interface for closing a runner function. Implement this for cleaning up things.
type Closer interface {
	Close(ctx context.Context) error
}

// TaskGroup is an asyncronous task runner that wraps a WaitGroup
type TaskGroup struct {
	Name string
	sync.WaitGroup
}

// NewTaskGroup creates new instance of TaskGroup
func NewTaskGroup(name string) *TaskGroup {
	return &TaskGroup{Name: name}
}

// Run executes a single asynchronous task.
//
// It creates a new trace for the task and passes through deadlines.
func (t *TaskGroup) Run(ctx context.Context, r Runner) {
	t.WaitGroup.Add(1)
	go func() {
		defer t.WaitGroup.Done()
		ctx2 := trace.StartSpan(ctx, t.Name)
		defer trace.End(ctx2)
		if err := r.Run(ctx2); err != nil && !errors.Is(err, context.Canceled) {
			log.Error(ctx2, t.Name, events.NewErrorInfo(err))
		}
	}()
}

// Loop repeatedly executes the provided task until it returns false
// or the context is canceled.
func (t *TaskGroup) Loop(ctx context.Context, r Runner) {
	run := func(ctx context.Context) bool {
		ctx2 := trace.StartSpan(ctx, t.Name)
		defer trace.End(ctx2)
		if err := r.Run(ctx2); err != nil && !errors.Is(err, context.Canceled) {
			log.Error(ctx2, t.Name, events.NewErrorInfo(err))
			return true
		}
		return false
	}

	t.WaitGroup.Add(1)
	go func() {
		defer t.WaitGroup.Done()
		for ctx.Err() == nil {
			if ok := run(ctx); ok {
				break
			}
		}
	}()
}

// Default is the default runner
var Default = NewTaskGroup("async.run")

// Run executes a single asynchronous task.
//
// It creates a new trace for the task and passes through deadlines.
func Run(ctx context.Context, r Runner) {
	Default.Run(ctx, r)
}

// RunClose closes any references a runner might be using
func RunClose(ctx context.Context, r Runner) error {
	switch r := r.(type) {
	case Closer:
		return r.Close(ctx)
	case io.Closer:
		return r.Close()
	}
	return nil
}

// RunGroup runs a group of runner tasks and exits when the first run group errors out
func RunGroup(rg []Runner) Runner {
	ru := Func(func(ctx context.Context) error {
		g, ctx := errgroup.WithContext(ctx)
		for idx := range rg {
			r := rg[idx]
			g.Go(func() error {
				defer func() {
					if err := RunClose(ctx, r); err != nil {
						log.Error(ctx, "Error when closing:", events.NewErrorInfo(err))
					}
				}()

				return r.Run(ctx)
			})
		}
		return g.Wait()
	})
	return ru
}

// Loop repeatedly executes the provided task until it returns false
// or the context is canceled.
func Loop(ctx context.Context, r Runner) {
	Default.Loop(ctx, r)
}

// Func is a helper that implements the Runner interface
type Func func(ctx context.Context) error

// Run implements the Runner interface
func (f Func) Run(ctx context.Context) error {
	return f(ctx)
}

// Sleep sleeps for the specified duration but can be canceled if the
// context is canceled or the context has an earlier deadline/timeout.
func Sleep(ctx context.Context, duration time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	<-ctx.Done()
}

// SleepUntil sleeps until the specified deadline but can be canceled
// if the context is canceled or has an earlier deadline.
func SleepUntil(ctx context.Context, deadline time.Time) {
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	<-ctx.Done()
}

// MutexWithContext is a lock that supports context cancellation.
//
// Note that unlike the `sync.Locker` style of mutex, this one's `Lock` method
// can fail and you must check its return value.
type MutexWithContext struct {
	sem *semaphore.Weighted
}

// NewMutexWithContext creates a new MutexWithContext instance.
func NewMutexWithContext() *MutexWithContext {
	return &MutexWithContext{semaphore.NewWeighted(1)}
}

// Lock acquires the mutex, blocking if it is unavailable.
//
// Unlike `sync.Mutex.Lock()`, this function can fail.  It is responsibility of
// the caller to check the returned error and not proceed if it is non-nil.
func (m *MutexWithContext) Lock(ctx context.Context) error {
	return m.sem.Acquire(ctx, 1)
}

// Unlock releases the mutex, allowing the next waiter to proceed.
func (m *MutexWithContext) Unlock() {
	m.sem.Release(1)
}
