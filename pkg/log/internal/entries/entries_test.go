//go:build !gobox_e2e

package entries_test

import (
	"testing"

	"github.com/grevych/gobox/pkg/log/internal/entries"
	"github.com/grevych/gobox/pkg/shuffler"
)

func TestAll(t *testing.T) {
	shuffler.Run(t, suite{})
}

type suite struct{}

func (suite) TestAppendNoBlock(t *testing.T) {
	items := entries.New()

	// fill the debug entry cache
	for i := 0; i < entries.MaxItems; i++ {
		items.Append("test")
	}

	ready := make(chan struct{}, 1)
	unblock := make(chan struct{})

	// start a blocking entry flush
	go func() {
		items.Flush(func(string) {
			select {
			// signal that the lock has been acquired and is now blocking.
			case ready <- struct{}{}:
			default:
			}
			<-unblock
		})
	}()
	<-ready

	// appends should continue to work without blocking
	for i := 0; i < entries.MaxItems; i++ {
		items.Append("test")
	}
	close(unblock)
}
