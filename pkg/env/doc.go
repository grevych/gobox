// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: provides environment specific overrides based on build tags

// Package env provides environment specific overrides
//
// All the functions provided by this package are meant to be called
// at app initialization and will effectively not do anything at all
// in production.
//
// This is done via build tags: gobox_test and gobox_dev represent the CI and
// dev-env environments.  The tags use the gobox_ prefix just in case
// some package in the dependency chain uses the same build tag to
// change their own behavior.
package env
