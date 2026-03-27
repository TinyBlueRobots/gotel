// Package gotel provides a unified initialization for all OpenTelemetry components.
// It simplifies setup by initializing tracing, metrics, and logging with a single call.
package gotel

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	shutdownMetrics, err := metrics.InitMetrics(ctx, serviceName, resourceAttrs, metricsStruct)
	if err != nil {
		shutdownErr := shutdownTracing(ctx)
		return nil, errors.Join(fmt.Errorf("failed to initialize metrics: %w", err), shutdownErr)
	}

	var shutdownLogger func(context.Context) error
	if logHandler != nil {
		shutdownLogger, err = log.InitLogger(ctx, resourceAttrs, logHandler)
	} else {
		shutdownLogger, err = log.InitLogger(ctx, resourceAttrs)
	}

	if err != nil {
		metricsErr := shutdownMetrics(ctx)
		tracingErr := shutdownTracing(ctx)

		return nil, errors.Join(fmt.Errorf("failed to initialize logger: %w", err), metricsErr, tracingErr)
	}

	shutdown := func(ctx context.Context) error {
		loggerErr := shutdownLogger(ctx)
		metricsErr := shutdownMetrics(ctx)
		tracingErr := shutdownTracing(ctx)

		return errors.Join(loggerErr, metricsErr, tracingErr)
	}

	return shutdown, nil
}
