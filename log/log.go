// Package log provides structured logging with automatic trace correlation.
// It integrates slog with OpenTelemetry for unified observability.
package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	slogmulti "github.com/samber/slog-multi"
	"github.com/tinybluerobots/gotel/attribute"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"go.opentelemetry.io/otel/trace"
)

type logWithContext func(ctx context.Context, message string, attributes ...attribute.Attr)

var noopLogWithContext = func(ctx context.Context, message string, attributes ...attribute.Attr) {}

var (
	// Debug logs a message at DEBUG level with optional attributes.
	Debug logWithContext = noopLogWithContext
	// Info logs a message at INFO level with optional attributes.
	Info logWithContext = noopLogWithContext
	// Warn logs a message at WARN level with optional attributes.
	Warn logWithContext = noopLogWithContext
	// Error logs an error at ERROR level with stack trace and optional attributes.
	Error func(ctx context.Context, err error, attributes ...attribute.Attr) = func(ctx context.Context, err error, attributes ...attribute.Attr) {}
)

func toSlogAttr(attr attribute.Attr) slog.Attr {
	key := string(attr.Key)
	value := attr.Value.AsInterface()

	return slog.Any(key, value)
}

// NewJSONHandler creates a JSON slog handler with resource attributes baked in.
func NewJSONHandler(w io.Writer, resourceAttrs []attribute.Attr, logLevel string) (slog.Handler, error) {
	slogResourceAttrs := make([]slog.Attr, len(resourceAttrs))

	for i, attr := range resourceAttrs {
		slogResourceAttrs[i] = slog.Attr{Key: string(attr.Key), Value: slog.AnyValue(attr.Value.AsInterface())}
	}

	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(logLevel)); err != nil {
		return nil, err
	}

	handlerOptions := &slog.HandlerOptions{Level: slogLevel}

	return slog.NewJSONHandler(w, handlerOptions).WithAttrs(slogResourceAttrs), nil
}

func newHttpLogger(ctx context.Context, insecure bool, resourceAttrs []attribute.Attr) (*log.LoggerProvider, error) {
	options := []otlploghttp.Option{}

	if insecure {
		options = append(options, otlploghttp.WithInsecure())
	}

	exp, err := otlploghttp.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exp)
	provider := log.NewLoggerProvider(log.WithProcessor(processor), log.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attribute.ToKeyValues(resourceAttrs)...)))

	return provider, nil
}

func newGrpcLogger(ctx context.Context, insecure bool, resourceAttrs []attribute.Attr) (*log.LoggerProvider, error) {
	options := []otlploggrpc.Option{}

	if insecure {
		options = append(options, otlploggrpc.WithInsecure())
	}

	exp, err := otlploggrpc.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	processor := log.NewBatchProcessor(exp)
	provider := log.NewLoggerProvider(log.WithProcessor(processor), log.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attribute.ToKeyValues(resourceAttrs)...)))

	return provider, nil
}

func grpcLogHandler(ctx context.Context, resourceAttrs []attribute.Attr) (slog.Handler, *log.LoggerProvider, error) {
	insecure := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true"
	useHttp := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "http"

	var (
		provider *log.LoggerProvider
		err      error
	)

	if useHttp {
		provider, err = newHttpLogger(ctx, insecure, resourceAttrs)
	} else {
		provider, err = newGrpcLogger(ctx, insecure, resourceAttrs)
	}

	if err != nil {
		return nil, nil, err
	}

	return otelslog.NewHandler("otelslog", otelslog.WithLoggerProvider(provider)), provider, nil
}

// InitLogger initializes structured logging with optional OTEL export.
// It sets up the package-level Debug, Info, Warn, and Error functions.
// Logs automatically include trace_id when within a valid trace context.
func InitLogger(ctx context.Context, resourceAttrs []attribute.Attr, handler ...slog.Handler) (func(context.Context) error, error) {
	slogHandlers := make([]slog.Handler, 0)
	slogHandlers = append(slogHandlers, handler...)

	var provider *log.LoggerProvider

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		otelHandler, loggerProvider, err := grpcLogHandler(ctx, resourceAttrs)
		if err != nil {
			return nil, err
		}

		slogHandlers = append(slogHandlers, otelHandler)
		provider = loggerProvider
	}

	fanoutHandler := slogmulti.Fanout(slogHandlers...)
	slogger := slog.New(fanoutHandler)

	writeLog := func(ctx context.Context, logF func(ctx context.Context, msg string, args ...any), message string, logAttributes ...attribute.Attr) {
		slogAttrs := make([]any, 0)
		for _, attribute := range logAttributes {
			slogAttrs = append(slogAttrs, toSlogAttr(attribute))
		}

		spanContext := trace.SpanFromContext(ctx).SpanContext()
		if spanContext.IsValid() {
			attr := slog.String("trace_id", spanContext.TraceID().String())
			slogAttrs = append(slogAttrs, attr)
		}

		logF(ctx, message, slogAttrs...)
	}

	Debug = func(ctx context.Context, message string, attributes ...attribute.Attr) {
		writeLog(ctx, slogger.DebugContext, message, attributes...)
	}
	Info = func(ctx context.Context, message string, attributes ...attribute.Attr) {
		writeLog(ctx, slogger.InfoContext, message, attributes...)
	}
	Warn = func(ctx context.Context, message string, attributes ...attribute.Attr) {
		writeLog(ctx, slogger.WarnContext, message, attributes...)
	}
	Error = func(ctx context.Context, err error, attributes ...attribute.Attr) {
		stackTrace := debug.Stack()
		attributes = append(attributes, attribute.New("stack_trace", string(stackTrace)))
		writeLog(ctx, slogger.ErrorContext, err.Error(), attributes...)
	}

	shutdown := func(ctx context.Context) error {
		if provider != nil {
			return provider.Shutdown(ctx)
		}

		return nil
	}

	return shutdown, nil
}
