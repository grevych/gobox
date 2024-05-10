//go:build !gobox_e2e

package events_test

import (
	"testing"

	"github.com/grevych/gobox/pkg/shuffler"
)

func TestAll(t *testing.T) {
	shuffler.Run(t, errorSuite{}, eventsSuite{})
}
