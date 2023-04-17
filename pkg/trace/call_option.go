// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Provides various call options functions

package trace

import (
	"context"
	"time"

	"github.com/getoutreach/gobox/internal/call"
)

// CallOptions contains options for all tracing calls. See
// call.Options for more information.
type CallOptions call.Options

// WithScheduledTime sets the call scheduled time to the provided
// time. Normally, this is set automatically by StartCall.
//
// Example:
//
//	ctx = trace.StartCall(ctx, "http", log.F{"query": query}, trace.WithScheduledTime(time.Now()))
func WithScheduledTime(t time.Time) call.Option {
	return func(c *call.Info) {
		c.Times.Scheduled = t
	}
}

// AsGRPCCall set the call type to GRPC
func AsGRPCCall() call.Option {
	return func(c *call.Info) {
		c.Type = call.TypeGRPC
	}
}

// AsHTTPCall set the call type to HTTP
func AsHTTPCall() call.Option {
	return func(c *call.Info) {
		c.Type = call.TypeHTTP
	}
}

// AsOutboundCall set the call type to Outbound
func AsOutboundCall() call.Option {
	return func(c *call.Info) {
		c.Type = call.TypeOutbound
	}
}

// SetCallOptions sets the provided call options on the current call in the
// provided context. The provided options replace any existing options.
// Call options are not preserved across application boundaries.
//
// Example:
//
//	ctx = trace.StartCall(ctx, "http", trace.WithCallOptions(ctx, trace.CallOptions{DisableInfoLogging: true}))
func WithCallOptions(ctx context.Context, opts CallOptions) {
	callTracker.Info(ctx).Opts = call.Options(opts)
}

// WithInfoLoggingDisabled disables info logging on the current call
//
// Example:
//
//	ctx = trace.StartCall(ctx, "http", trace.WithInfoLoggingDisabled())
func WithInfoLoggingDisabled() call.Option {
	return func(c *call.Info) {
		c.Opts.DisableInfoLogging = true
	}
}
