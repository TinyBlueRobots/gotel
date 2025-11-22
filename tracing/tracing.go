// Package tracing provides simplified distributed tracing with OTLP exporters.
// It wraps OpenTelemetry spans with a streamlined API for common operations.
package tracing

import (
	"context"
	"os"

	"github.com/tinybluerobots/gotel/attribute"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// StatusCode represents the status of a span.
type StatusCode codes.Code

const (
	// StatusUnset is the default status.
	StatusUnset StatusCode = StatusCode(codes.Unset)
	// StatusError indicates the operation failed.
	StatusError StatusCode = StatusCode(codes.Error)
	// StatusOk indicates the operation completed successfully.
	StatusOk StatusCode = StatusCode(codes.Ok)
)

// Span wraps an OpenTelemetry span with a simplified API.
type Span struct {
	traceSpan trace.Span
}

// AddEvent adds an event to the span with optional attributes.
func (s *Span) AddEvent(name string, attrs ...attribute.Attr) {
	otelAttrs := make([]otelattribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = attr.KeyValue
	}

	s.traceSpan.AddEvent(name, trace.WithAttributes(otelAttrs...))
}

// RecordError records an error on the span without setting status.
func (s *Span) RecordError(err error) {
	s.traceSpan.RecordError(err)
}

// RecordErrorAndSetStatus records an error and sets the span status to Error.
func (s *Span) RecordErrorAndSetStatus(err error) {
	s.RecordError(err)
	s.traceSpan.SetStatus(codes.Error, err.Error())
}

// SetStatus sets the span status with a code and description.
func (s *Span) SetStatus(code StatusCode, description string) {
	s.traceSpan.SetStatus(codes.Code(code), description)
}

// SetOk sets the span status to Ok.
func (s *Span) SetOk() {
	s.traceSpan.SetStatus(codes.Ok, "")
}

// SetAttributes sets attributes on the span.
func (s *Span) SetAttributes(attrs ...attribute.Attr) {
	otelAttrs := make([]otelattribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = attr.KeyValue
	}

	s.traceSpan.SetAttributes(otelAttrs...)
}

// End completes the span.
func (s *Span) End() {
	s.traceSpan.End()
}

var tracer = noop.NewTracerProvider().Tracer("noop")

func init() {
	otel.SetTextMapPropagator(propagation.TraceContext{})
}

func newGrpcTraceExporter(ctx context.Context, insecure bool) (sdktrace.SpanExporter, error) {
	options := []otlptracegrpc.Option{}

	if insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(ctx, options...)
}

func newHttpTraceExporter(ctx context.Context, insecure bool) (sdktrace.SpanExporter, error) {
	options := []otlptracehttp.Option{}

	if insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}

	return otlptracehttp.New(ctx, options...)
}

// InitTracing initializes the tracer with OTLP exporters.
// Returns a shutdown function to flush and close the tracer provider.
func InitTracing(ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, options ...sdktrace.TracerProviderOption) (func(context.Context) error, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		insecure := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true"
		useHttp := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "http"

		var (
			exporter sdktrace.SpanExporter
			err      error
		)

		if useHttp {
			exporter, err = newHttpTraceExporter(ctx, insecure)
		} else {
			exporter, err = newGrpcTraceExporter(ctx, insecure)
		}

		if err != nil {
			return nil, err
		}

		options = append(options, sdktrace.WithBatcher(exporter))
	}

	options = append(options, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attribute.ToKeyValues(resourceAttrs)...)))
	provider := sdktrace.NewTracerProvider(options...)
	tracer = provider.Tracer(serviceName)

	return provider.Shutdown, nil
}

// TraceHeaders extracts W3C trace context headers for propagation to downstream services.
func TraceHeaders(ctx context.Context) map[string]string {
	metadata := map[string]string{}
	carrier := propagation.MapCarrier(metadata)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return metadata
}

func newSpan(ctx context.Context, name string, attrs ...attribute.Attr) (context.Context, Span) {
	otelAttrs := make([]otelattribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = attr.KeyValue
	}

	ctx, traceSpan := tracer.Start(ctx, name, trace.WithAttributes(otelAttrs...))

	return ctx, Span{traceSpan}
}

// NewSpan creates a new span with the given name and optional attributes.
func NewSpan(ctx context.Context, name string, attrs ...attribute.Attr) (context.Context, Span) {
	return newSpan(ctx, name, attrs...)
}

// NewChildSpan creates a child span from propagated trace context headers.
func NewChildSpan(ctx context.Context, carrier map[string]string, name string, attrs ...attribute.Attr) (context.Context, Span) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(carrier))
	return newSpan(ctx, name, attrs...)
}
