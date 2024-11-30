package loglevelswitcher

import (
	"context"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
)

func TestServiceActivity_RunAndClose(t *testing.T) {
	var loglevelswitcherErr error
	wg := sync.WaitGroup{}

	log := logrus.New()
	loglevelswitcherSvc := New(log, syscall.SIGHUP)

	// Make sure default level is INFO
	assert.Assert(t, log.GetLevel() == logrus.InfoLevel)

	wg.Add(1)
	go func() {
		defer wg.Done()
		loglevelswitcherErr = loglevelswitcherSvc.Run(context.Background())
	}()

	// Wait for service activity to get started
	time.Sleep(200 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	// Wait for the service activity to assign level
	time.Sleep(200 * time.Millisecond)

	assert.Assert(t, log.GetLevel() == logrus.DebugLevel)

	loglevelswitcherSvc.Close(context.Background())
	wg.Wait()

	assert.NilError(t, loglevelswitcherErr)
}

func TestServiceActivity_RunAndCancelContext(t *testing.T) {
	var loglevelswitcherErr error
	wg := sync.WaitGroup{}

	log := logrus.New()
	loglevelswitcherSvc := New(log, syscall.SIGHUP)

	// Make sure default level is INFO
	assert.Assert(t, log.GetLevel() == logrus.InfoLevel)

	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		loglevelswitcherErr = loglevelswitcherSvc.Run(ctx)
	}()

	// Wait for service activity to get started
	time.Sleep(200 * time.Millisecond)
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
	// Wait for the service activity to assign level
	time.Sleep(200 * time.Millisecond)

	assert.Assert(t, log.GetLevel() == logrus.DebugLevel)

	cancel()
	wg.Wait()

	assert.ErrorContains(t, loglevelswitcherErr, "context canceled")
}
