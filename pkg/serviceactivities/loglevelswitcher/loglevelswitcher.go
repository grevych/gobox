package loglevelswitcher

import (
	"context"
	"os"
	"os/signal"

	"github.com/grevych/gobox/pkg/async"
	"github.com/sirupsen/logrus"
)

// Logger is an interface used by the log level switcher Service Activity
// to set the level of a logger.
type Logger interface {
	SetLevel(logrus.Level) // inside this function you can log the change of level
}

// ServiceActivity implements the async.Runner & async.Closer interface for
// switching a logger level through command line.
type ServiceActivity struct {
	signal       os.Signal
	logger       Logger
	done         chan struct{}
	debugLevelOn bool
}

// Make sure ServiceActivity implements async.Runner interface.
var _ async.Runner = (*ServiceActivity)(nil)

// Make sure ServiceActivity implemets async.Closer interface.
var _ async.Closer = (*ServiceActivity)(nil)

// New creates a new service activity that listens for a specific os signal
// and attempts to toggle the level of the provided logger to debug mode.
func New(l Logger, s os.Signal) *ServiceActivity {
	return &ServiceActivity{
		signal:       s,
		logger:       l,
		debugLevelOn: false,
		done:         make(chan struct{}),
	}
}

// Run runs the log level switcher service activity
func (sa *ServiceActivity) Run(ctx context.Context) error {
	// listen for the given signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, sa.signal)

	for {
		select {
		case <-c:
			var logLevel logrus.Level
			if !sa.debugLevelOn {
				logLevel = logrus.DebugLevel
			} else {
				logLevel = logrus.InfoLevel
			}
			sa.debugLevelOn = !sa.debugLevelOn
			sa.logger.SetLevel(logLevel)
		case <-ctx.Done():
			return ctx.Err()
		case <-sa.done:
			return nil
		}
	}
}

// Close closes the log level switcher service activity
func (s *ServiceActivity) Close(_ context.Context) error {
	close(s.done)
	return nil
}
