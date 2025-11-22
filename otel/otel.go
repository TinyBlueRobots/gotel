// Package otel provides a unified initialization for all OpenTelemetry components.
// It simplifies setup by initializing tracing, metrics, and logging with a single call.
package otel

import (
	"context"
	"io"
	"os"

	"github.com/tinybluerobots/gotel/attribute"
	"github.com/tinybluerobots/gotel/log"
	"github.com/tinybluerobots/gotel/metrics"
	"github.com/tinybluerobots/gotel/tracing"
)

type options struct {
	logWriter io.Writer
	logLevel  string
}

// Option configures the Init function.
type Option func(*options)

// WithLogWriter sets the io.Writer for log output. Default is os.Stdout.
func WithLogWriter(w io.Writer) Option {
	return func(o *options) {
		o.logWriter = w
	}
}

// WithLogLevel sets the log level. Default is "INFO".
// Valid levels: DEBUG, INFO, WARN, ERROR.
func WithLogLevel(level string) Option {
	return func(o *options) {
		o.logLevel = level
	}
}

// Init initializes all telemetry components (tracing, metrics, logging) with a single call.
// Returns a shutdown function that gracefully closes all providers.
func Init[T any](ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, metricsStruct *T, opts ...Option) (func(context.Context) error, error) {
	o := &options{
		logWriter: os.Stdout,
		logLevel:  "INFO",
	}

	for _, opt := range opts {
		opt(o)
	}

	shutdownTracing, err := tracing.InitTracing(ctx, serviceName, resourceAttrs)
	if err != nil {
		return nil, err
	}

	shutdownMetrics, err := metrics.InitMetrics(ctx, serviceName, resourceAttrs, metricsStruct)
	if err != nil {
		_ = shutdownTracing(ctx)
		return nil, err
	}

	logHandler, err := log.NewJSONHandler(o.logWriter, resourceAttrs, o.logLevel)
	if err != nil {
		_ = shutdownMetrics(ctx)
		_ = shutdownTracing(ctx)

		return nil, err
	}

	shutdownLogger, err := log.InitLogger(ctx, resourceAttrs, logHandler)
	if err != nil {
		_ = shutdownMetrics(ctx)
		_ = shutdownTracing(ctx)

		return nil, err
	}

	shutdown := func(ctx context.Context) error {
		var firstErr error
		if err := shutdownLogger(ctx); err != nil && firstErr == nil {
			firstErr = err
		}

		if err := shutdownMetrics(ctx); err != nil && firstErr == nil {
			firstErr = err
		}

		if err := shutdownTracing(ctx); err != nil && firstErr == nil {
			firstErr = err
		}

		return firstErr
	}

	return shutdown, nil
}
