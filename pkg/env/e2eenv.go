// Copyright 2022 Outreach Corporation. All Rights Reserved.

//go:build gobox_e2e
// +build gobox_e2e

// Description: Provides environment overrides for e2e tests

package env

import "github.com/grevych/gobox/pkg/cfg"

func ApplyOverrides() {
	old := cfg.DefaultReader()
	cfg.SetDefaultReader(testReader(devReader(old), &overrides))
}

func init() { //nolint:gochecknoinits // Why: On purpose.
	ApplyOverrides()
}
