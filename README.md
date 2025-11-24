# gotel

A Go library providing simplified wrappers around OpenTelemetry SDK components for distributed tracing, metrics collection, and structured logging.

## Installation

```bash
go get github.com/tinybluerobots/gotel
```

## Quick Start

```go
package main

import (
    "context"
    "os"

    "github.com/tinybluerobots/gotel"
    "github.com/tinybluerobots/gotel/attribute"
    "github.com/tinybluerobots/gotel/log"
    "github.com/tinybluerobots/gotel/metrics"
    "github.com/tinybluerobots/gotel/tracing"
)

type AppMetrics struct {
    RequestCount *metrics.Int64Counter
}

func main() {
    ctx := context.Background()

    // Create resource attributes
    resourceAttrs := attribute.ResourceAttributes("myservice", "1.0.0", "production", "myhost")

    // Create a JSON log handler for stdout logging
    logHandler, err := log.NewJSONHandler(os.Stdout, resourceAttrs, "INFO")
    if err != nil {
        panic(err)
    }

    // Initialize all telemetry (tracing, metrics, logging)
    shutdown, err := gotel.Init(ctx, "myservice", resourceAttrs, &AppMetrics{}, logHandler)
    if err != nil {
        panic(err)
    }
    defer shutdown(ctx)

    // Create a span
    ctx, span := tracing.NewSpan(ctx, "operation",
        attribute.New("user_id", "123"))
    defer span.End()

    // Record a metric
    m := metrics.Metrics[AppMetrics]()
    m.RequestCount.Add(ctx, 1, attribute.New("method", "GET"))

    // Log with trace correlation
    log.Info(ctx, "Operation started", attribute.New("key", "value"))
}
```

## Configuration

The library uses standard OpenTelemetry environment variables:

| Variable | Description | Values |
|----------|-------------|--------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP backend endpoint | URL (e.g., `http://localhost:4317`) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Export protocol | `grpc` (default), `http` |
| `OTEL_EXPORTER_OTLP_INSECURE` | Disable TLS | `true`, `false` (default) |

Exporters are only created when `OTEL_EXPORTER_OTLP_ENDPOINT` is set.

## API Reference

### Unified Initialization

#### Init

Initialize all telemetry components (tracing, metrics, logging) with a single call.

```go
func Init[T any](ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, metricsStruct *T, logHandler slog.Handler) (func(context.Context) error, error)
```

Pass a `slog.Handler` to enable local logging (use `log.NewJSONHandler`), or `nil` to log only to the OTEL collector.

```go
import "github.com/tinybluerobots/gotel"

shutdown, err := gotel.Init(ctx, "myservice", resourceAttrs, &AppMetrics{}, logHandler)
```

### Tracing

#### InitTracing

Initialize the tracer with OTLP exporters.

```go
func InitTracing(ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, options ...sdktrace.TracerProviderOption) (func(context.Context) error, error)
```

#### NewSpan

Create a new top-level span.

```go
func NewSpan(ctx context.Context, name string, attrs ...attribute.Attr) (context.Context, tracing.Span)
```

#### TraceHeaders

Extract W3C trace context headers for propagation.

```go
func TraceHeaders(ctx context.Context) map[string]string
```

### Span

The `tracing.Span` type wraps OpenTelemetry spans with a simplified interface.

```go
// Add an event to the span
span.AddEvent(name string, attrs ...attribute.Attr)

// Record an error without setting status
span.RecordError(err error)

// Record error and set span status to Error
span.RecordErrorAndSetStatus(err error)

// Set span status
span.SetStatus(code tracing.StatusCode, description string)

// Set span status to Ok
span.SetOk()

// Set span attributes
span.SetAttributes(attrs ...attribute.Attr)

// End the span
span.End()
```

### Metrics

#### InitMetrics

Initialize metrics with OTLP exporters. Metric instruments are registered via reflection on the provided struct.

```go
func InitMetrics[T any](ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, m *T, options ...sdkmetric.Option) (func(context.Context) error, error)
```

#### Metrics

Retrieve the initialized metrics struct. Returns nil if not initialized.

```go
func Metrics[T any]() *T
```

#### Usage Example

```go
type MyMetrics struct {
    RequestCount *metrics.Int64Counter
    ResponseTime *metrics.Int64Histogram
    ActiveUsers  *metrics.Int64Gauge
}

// Initialize
resourceAttrs := attribute.ResourceAttributes("myservice", "1.0.0", "production", "myhost")
metrics.InitMetrics(ctx, "myservice", resourceAttrs, &MyMetrics{})

// Use
m := metrics.Metrics[MyMetrics]()
m.RequestCount.Add(ctx, 1, attribute.New("method", "GET"))
m.ResponseTime.Record(ctx, 150, attribute.New("endpoint", "/api"))
m.ActiveUsers.Record(ctx, 42)
```

#### Metric Types

**Counters** (monotonically increasing):
- `*metrics.Int64Counter` - `Add(ctx, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64Counter` - `Add(ctx, value float64, attrs ...attribute.Attr)`

**Up/Down Counters** (can increase or decrease):
- `*metrics.Int64UpDownCounter` - `Add(ctx, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64UpDownCounter` - `Add(ctx, value float64, attrs ...attribute.Attr)`

**Gauges** (instantaneous measurements):
- `*metrics.Int64Gauge` - `Record(ctx, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64Gauge` - `Record(ctx, value float64, attrs ...attribute.Attr)`

**Histograms** (distribution of values):
- `*metrics.Int64Histogram` - `Record(ctx, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64Histogram` - `Record(ctx, value float64, attrs ...attribute.Attr)`

**Observable Counters** (callback-based):
- `*metrics.Int64ObservableCounter` - `Observe(observer, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64ObservableCounter` - `Observe(observer, value float64, attrs ...attribute.Attr)`

**Observable Up/Down Counters** (callback-based):
- `*metrics.Int64ObservableUpDownCounter` - `Observe(observer, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64ObservableUpDownCounter` - `Observe(observer, value float64, attrs ...attribute.Attr)`

**Observable Gauges** (callback-based):
- `*metrics.Int64ObservableGauge` - `Observe(observer, value int64, attrs ...attribute.Attr)`
- `*metrics.Float64ObservableGauge` - `Observe(observer, value float64, attrs ...attribute.Attr)`

### Logging

#### NewJSONHandler

Create a JSON slog handler with resource attributes baked in. Pass this to `otel.Init` via `otel.WithLogHandler` to enable logging.

```go
func NewJSONHandler(w io.Writer, resourceAttrs []attribute.Attr, logLevel string) (slog.Handler, error)
```

#### InitLogger

Initialize structured logging with slog and optional OTEL exporter.

```go
func InitLogger(ctx context.Context, resourceAttrs []attribute.Attr, handler ...slog.Handler) (func(context.Context) error, error)
```

Log levels: `DEBUG`, `INFO`, `WARN`, `ERROR`

#### Log Functions

```go
log.Debug(ctx context.Context, message string, attributes ...attribute.Attr)
log.Info(ctx context.Context, message string, attributes ...attribute.Attr)
log.Warn(ctx context.Context, message string, attributes ...attribute.Attr)
log.Error(ctx context.Context, err error, attributes ...attribute.Attr)
```

Logs automatically include trace IDs when within a valid trace context. Error logging captures stack traces.

### Attributes

#### New

Create an attribute with automatic type conversion.

```go
func New(key string, value any) attribute.Attr
```

Supported types:
- `bool`, `[]bool`
- `float64`, `[]float64`
- `int`, `[]int`
- `int64`, `[]int64`
- `string`, `[]string`
- `fmt.Stringer` (converted to string)
- Any other type (formatted with `%v`)

## Complete Example

```go
package main

import (
    "context"
    "errors"
    "os"

    "github.com/tinybluerobots/gotel"
    "github.com/tinybluerobots/gotel/attribute"
    "github.com/tinybluerobots/gotel/log"
    "github.com/tinybluerobots/gotel/metrics"
    "github.com/tinybluerobots/gotel/tracing"
)

type AppMetrics struct {
    RequestCount *metrics.Int64Counter
    RequestTime  *metrics.Int64Histogram
}

func main() {
    ctx := context.Background()

    // Create resource attributes
    resourceAttrs := attribute.ResourceAttributes("myapp", "1.0.0", "production", "myhost")

    // Create a JSON log handler
    logHandler, err := log.NewJSONHandler(os.Stdout, resourceAttrs, "INFO")
    if err != nil {
        panic(err)
    }

    // Initialize all telemetry
    shutdown, err := gotel.Init(ctx, "myapp", resourceAttrs, &AppMetrics{}, logHandler)
    if err != nil {
        panic(err)
    }
    defer shutdown(ctx)

    // Get metrics
    m := metrics.Metrics[AppMetrics]()

    // Start a span
    ctx, span := tracing.NewSpan(ctx, "handleRequest",
        attribute.New("method", "GET"),
        attribute.New("path", "/api/users"))
    defer span.End()

    // Log with trace correlation
    log.Info(ctx, "Processing request")

    // Record metrics
    m.RequestCount.Add(ctx, 1, attribute.New("method", "GET"))

    // Simulate error
    err = errors.New("database connection failed")
    if err != nil {
        span.RecordErrorAndSetStatus(err)
        log.Error(ctx, err, attribute.New("operation", "fetchData"))
    }

    // Mark span as successful
    span.SetOk()

    // Extract headers for propagation to downstream services
    headers := tracing.TraceHeaders(ctx)
    _ = headers // Use for outgoing HTTP requests

    m.RequestTime.Record(ctx, 150, attribute.New("method", "GET"))
}
```
