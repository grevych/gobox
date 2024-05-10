//go:build !gobox_e2e

package log_test

import (
	"testing"

	"github.com/grevych/gobox/pkg/shuffler"
)

func TestAll(t *testing.T) {
	shuffler.Run(t, fatalSuite{}, withSuite{}, callerSuite{})
}
