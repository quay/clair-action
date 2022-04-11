// Package zlog is a logging facade backed by zerolog.
//
// It uses opentelemetry baggage to generate log contexts.
//
// By default, the package wraps the zerolog global logger. This can be changed
// via the Set function.
//
// In addition, a testing adapter is provided to keep testing logs orderly.
package zlog

import (
	"context"

	"github.com/rs/zerolog"
	global "github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/baggage"
)

// Log is the logger used by the package-level functions.
var log = &global.Logger

// Set configures the logger used by this package.
//
// This function is unsafe to use concurrently with the other functions of this
// package.
func Set(l *zerolog.Logger) {
	log = l
}

// AddCtx is the workhorse function that every facade function calls.
//
// If the passed Event is enabled, it will attach all the otel baggage to
// it and return it.
func addCtx(ctx context.Context, ev *zerolog.Event) *zerolog.Event {
	if !ev.Enabled() {
		return ev
	}

	b := baggage.FromContext(ctx)
	for _, m := range b.Members() {
		ev.Str(m.Key(), m.Value())
	}

	return ev
}

// Log starts a new message with no level.
func Log(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Log())
}

// WithLevel starts a new message with the specified level.
func WithLevel(ctx context.Context, l zerolog.Level) *zerolog.Event {
	return addCtx(ctx, log.WithLevel(l))
}

// Trace starts a new message with the trace level.
func Trace(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Trace())
}

// Debug starts a new message with the debug level.
func Debug(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Debug())
}

// Info starts a new message with the infor level.
func Info(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Info())
}

// Warn starts a new message with the warn level.
func Warn(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Warn())
}

// Error starts a new message with the error level.
func Error(ctx context.Context) *zerolog.Event {
	return addCtx(ctx, log.Error())
}
