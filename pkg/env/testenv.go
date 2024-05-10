//go:build gobox_test && !gobox_e2e
// +build gobox_test,!gobox_e2e

package env

import (
	"github.com/grevych/gobox/pkg/cfg"
)

func ApplyOverrides() {
	old := cfg.DefaultReader()
	cfg.SetDefaultReader(testReader(old, &overrides))
}

func init() { //nolint: gochecknoinits
	ApplyOverrides()
}
