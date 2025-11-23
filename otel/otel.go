// Package otel provides a unified initialization for all OpenTelemetry components.
// It simplifies setup by initializing tracing, metrics, and logging with a single call.
package otel

import (
	"context"
	"log/slog"

	"github.com/tinybluerobots/gotel/attribute"
	"github.com/tinybluerobots/gotel/log"
	"github.com/tinybluerobots/gotel/metrics"
	"github.com/tinybluerobots/gotel/tracing"
)

// Init initializes all telemetry components (tracing, metrics, logging) with a single call.
// Returns a shutdown function that gracefully closes all providers.
// Pass a slog.Handler to enable local logging, or nil to log only to the OTEL collector.
func Init[T any](ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, metricsStruct *T, logHandler slog.Handler) (func(context.Context) error, error) {
	shutdownTracing, err := tracing.InitTracing(ctx, serviceName, resourceAttrs)
	if err != nil {
		return nil, err
	}

	shutdownMetrics, err := metrics.InitMetrics(ctx, serviceName, resourceAttrs, metricsStruct)
	if err != nil {
		_ = shutdownTracing(ctx)
		return nil, err
	}

	var shutdownLogger func(context.Context) error
	if logHandler != nil {
		shutdownLogger, err = log.InitLogger(ctx, resourceAttrs, logHandler)
	} else {
		shutdownLogger, err = log.InitLogger(ctx, resourceAttrs)
	}

	if err != nil {
		_ = shutdownMetrics(ctx)
		_ = shutdownTracing(ctx)

		return nil, err
	}

	shutdown := func(ctx context.Context) error {
		firstErr := shutdownLogger(ctx)

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
