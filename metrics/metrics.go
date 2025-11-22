// Package metrics provides simplified metrics collection with OTLP exporters.
// It uses reflection to automatically initialize metric instruments from struct fields.
package metrics

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/tinybluerobots/gotel/attribute"
	otelattribute "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

var metricsInstance any

// Metrics retrieves the initialized metrics struct.
// Returns nil if metrics have not been initialized or if the type doesn't match.
func Metrics[T any]() *T {
	if metricsInstance == nil {
		return nil
	}

	m, ok := metricsInstance.(*T)
	if !ok {
		return nil
	}

	return m
}

// Int64Counter is a monotonically increasing counter for int64 values.
type Int64Counter struct {
	int64Counter metric.Int64Counter
}

// Float64Counter is a monotonically increasing counter for float64 values.
type Float64Counter struct {
	float64Counter metric.Float64Counter
}

// Int64UpDownCounter is a counter that can increase or decrease for int64 values.
type Int64UpDownCounter struct {
	int64UpDownCounter metric.Int64UpDownCounter
}

// Float64UpDownCounter is a counter that can increase or decrease for float64 values.
type Float64UpDownCounter struct {
	float64UpDownCounter metric.Float64UpDownCounter
}

// Int64ObservableCounter is a callback-based monotonically increasing counter for int64 values.
type Int64ObservableCounter struct {
	int64ObservableCounter metric.Int64ObservableCounter
	meter                  metric.Meter
}

// Float64ObservableCounter is a callback-based monotonically increasing counter for float64 values.
type Float64ObservableCounter struct {
	float64ObservableCounter metric.Float64ObservableCounter
	meter                    metric.Meter
}

// Int64ObservableUpDownCounter is a callback-based counter that can increase or decrease for int64 values.
type Int64ObservableUpDownCounter struct {
	int64ObservableUpDownCounter metric.Int64ObservableUpDownCounter
	meter                        metric.Meter
}

// Float64ObservableUpDownCounter is a callback-based counter that can increase or decrease for float64 values.
type Float64ObservableUpDownCounter struct {
	float64ObservableUpDownCounter metric.Float64ObservableUpDownCounter
	meter                          metric.Meter
}

// Int64Gauge records instantaneous int64 measurements.
type Int64Gauge struct {
	int64Gauge metric.Int64Gauge
}

// Float64Gauge records instantaneous float64 measurements.
type Float64Gauge struct {
	float64Gauge metric.Float64Gauge
}

// Int64ObservableGauge is a callback-based gauge for int64 values.
type Int64ObservableGauge struct {
	int64ObservableGauge metric.Int64ObservableGauge
	meter                metric.Meter
}

// Float64ObservableGauge is a callback-based gauge for float64 values.
type Float64ObservableGauge struct {
	float64ObservableGauge metric.Float64ObservableGauge
	meter                  metric.Meter
}

// Int64Histogram records a distribution of int64 values.
type Int64Histogram struct {
	int64Histogram metric.Int64Histogram
}

// Float64Histogram records a distribution of float64 values.
type Float64Histogram struct {
	float64Histogram metric.Float64Histogram
}

func newAttributeSet(attrs ...attribute.Attr) otelattribute.Set {
	otelAttrs := make([]otelattribute.KeyValue, len(attrs))

	for i, attr := range attrs {
		otelAttrs[i] = attr.KeyValue
	}

	return otelattribute.NewSet(otelAttrs...)
}

// Add increments the counter by the given value.
func (c *Int64Counter) Add(ctx context.Context, Value int64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		c.int64Counter.Add(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Add increments the counter by the given value.
func (c *Float64Counter) Add(ctx context.Context, Value float64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		c.float64Counter.Add(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Add adds the given value to the counter (can be negative).
func (c *Int64UpDownCounter) Add(ctx context.Context, Value int64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		c.int64UpDownCounter.Add(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Add adds the given value to the counter (can be negative).
func (c *Float64UpDownCounter) Add(ctx context.Context, Value float64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		c.float64UpDownCounter.Add(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Record records a measurement.
func (g *Int64Gauge) Record(ctx context.Context, Value int64, attrs ...attribute.Attr) {
	if g != nil {
		attributeSet := newAttributeSet(attrs...)
		g.int64Gauge.Record(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Record records a measurement.
func (g *Float64Gauge) Record(ctx context.Context, Value float64, attrs ...attribute.Attr) {
	if g != nil {
		attributeSet := newAttributeSet(attrs...)
		g.float64Gauge.Record(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Record records a value in the histogram distribution.
func (h *Int64Histogram) Record(ctx context.Context, Value int64, attrs ...attribute.Attr) {
	if h != nil {
		attributeSet := newAttributeSet(attrs...)
		h.int64Histogram.Record(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Record records a value in the histogram distribution.
func (h *Float64Histogram) Record(ctx context.Context, Value float64, attrs ...attribute.Attr) {
	if h != nil {
		attributeSet := newAttributeSet(attrs...)
		h.float64Histogram.Record(ctx, Value, metric.WithAttributeSet(attributeSet))
	}
}

// Observe records a value from within a callback.
func (c *Int64ObservableCounter) Observe(observer metric.Int64Observer, value int64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (c *Int64ObservableCounter) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if c == nil {
		return nil, nil
	}

	return c.meter.RegisterCallback(callback, c.int64ObservableCounter)
}

// Observe records a value from within a callback.
func (c *Float64ObservableCounter) Observe(observer metric.Float64Observer, value float64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (c *Float64ObservableCounter) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if c == nil {
		return nil, nil
	}

	return c.meter.RegisterCallback(callback, c.float64ObservableCounter)
}

// Observe records a value from within a callback.
func (c *Int64ObservableUpDownCounter) Observe(observer metric.Int64Observer, value int64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (c *Int64ObservableUpDownCounter) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if c == nil {
		return nil, nil
	}

	return c.meter.RegisterCallback(callback, c.int64ObservableUpDownCounter)
}

// Observe records a value from within a callback.
func (c *Float64ObservableUpDownCounter) Observe(observer metric.Float64Observer, value float64, attrs ...attribute.Attr) {
	if c != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (c *Float64ObservableUpDownCounter) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if c == nil {
		return nil, nil
	}

	return c.meter.RegisterCallback(callback, c.float64ObservableUpDownCounter)
}

// Observe records a value from within a callback.
func (g *Int64ObservableGauge) Observe(observer metric.Int64Observer, value int64, attrs ...attribute.Attr) {
	if g != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (g *Int64ObservableGauge) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if g == nil {
		return nil, nil
	}

	return g.meter.RegisterCallback(callback, g.int64ObservableGauge)
}

// Observe records a value from within a callback.
func (g *Float64ObservableGauge) Observe(observer metric.Float64Observer, value float64, attrs ...attribute.Attr) {
	if g != nil {
		attributeSet := newAttributeSet(attrs...)
		observer.Observe(value, metric.WithAttributeSet(attributeSet))
	}
}

// RegisterCallback registers a callback to observe values.
func (g *Float64ObservableGauge) RegisterCallback(callback func(ctx context.Context, o metric.Observer) error) (metric.Registration, error) {
	if g == nil {
		return nil, nil
	}

	return g.meter.RegisterCallback(callback, g.float64ObservableGauge)
}

func newInstrument[T any, U any](name string, newInstrument func(string, ...U) (T, error)) (T, error) {
	c, err := newInstrument(name)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to create metric instrument %s: %w", name, err)
	}

	return c, nil
}

func initMetricFields(meter metric.Meter, m any) error {
	if m == nil || reflect.ValueOf(m).IsNil() {
		return nil
	}

	metricsInstance = m

	v := reflect.ValueOf(m).Elem()
	for i := range v.NumField() {
		field := v.Field(i)

		fieldName := v.Type().Field(i).Name
		switch field.Type() {
		case reflect.TypeOf(&Int64Counter{}):
			inst, err := newInstrument(fieldName, meter.Int64Counter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64Counter{inst}))
		case reflect.TypeOf(&Float64Counter{}):
			inst, err := newInstrument(fieldName, meter.Float64Counter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64Counter{inst}))
		case reflect.TypeOf(&Int64UpDownCounter{}):
			inst, err := newInstrument(fieldName, meter.Int64UpDownCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64UpDownCounter{inst}))
		case reflect.TypeOf(&Float64UpDownCounter{}):
			inst, err := newInstrument(fieldName, meter.Float64UpDownCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64UpDownCounter{inst}))
		case reflect.TypeOf(&Int64ObservableCounter{}):
			inst, err := newInstrument(fieldName, meter.Int64ObservableCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64ObservableCounter{inst, meter}))
		case reflect.TypeOf(&Float64ObservableCounter{}):
			inst, err := newInstrument(fieldName, meter.Float64ObservableCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64ObservableCounter{inst, meter}))
		case reflect.TypeOf(&Int64ObservableUpDownCounter{}):
			inst, err := newInstrument(fieldName, meter.Int64ObservableUpDownCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64ObservableUpDownCounter{inst, meter}))
		case reflect.TypeOf(&Float64ObservableUpDownCounter{}):
			inst, err := newInstrument(fieldName, meter.Float64ObservableUpDownCounter)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64ObservableUpDownCounter{inst, meter}))
		case reflect.TypeOf(&Int64Gauge{}):
			inst, err := newInstrument(fieldName, meter.Int64Gauge)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64Gauge{inst}))
		case reflect.TypeOf(&Float64Gauge{}):
			inst, err := newInstrument(fieldName, meter.Float64Gauge)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64Gauge{inst}))
		case reflect.TypeOf(&Int64ObservableGauge{}):
			inst, err := newInstrument(fieldName, meter.Int64ObservableGauge)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64ObservableGauge{inst, meter}))
		case reflect.TypeOf(&Float64ObservableGauge{}):
			inst, err := newInstrument(fieldName, meter.Float64ObservableGauge)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64ObservableGauge{inst, meter}))
		case reflect.TypeOf(&Int64Histogram{}):
			inst, err := newInstrument(fieldName, meter.Int64Histogram)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Int64Histogram{inst}))
		case reflect.TypeOf(&Float64Histogram{}):
			inst, err := newInstrument(fieldName, meter.Float64Histogram)
			if err != nil {
				return err
			}

			field.Set(reflect.ValueOf(&Float64Histogram{inst}))
		}
	}

	return nil
}

func newGrpcMetricExporter(ctx context.Context, insecure bool) (sdkmetric.Exporter, error) {
	options := []otlpmetricgrpc.Option{}

	if insecure {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}

	return otlpmetricgrpc.New(ctx, options...)
}

func newHttpMetricExporter(ctx context.Context, insecure bool) (sdkmetric.Exporter, error) {
	options := []otlpmetrichttp.Option{}

	if insecure {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	return otlpmetrichttp.New(ctx, options...)
}

// InitMetrics initializes metrics with OTLP exporters.
// Metric instruments are automatically created from the struct fields using reflection.
// Returns a shutdown function to flush and close the meter provider.
func InitMetrics[T any](ctx context.Context, serviceName string, resourceAttrs []attribute.Attr, metricsStruct *T, options ...sdkmetric.Option) (func(context.Context) error, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		insecure := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE") == "true"
		useHttp := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") == "http"

		var (
			exporter sdkmetric.Exporter
			err      error
		)

		if useHttp {
			exporter, err = newHttpMetricExporter(ctx, insecure)
		} else {
			exporter, err = newGrpcMetricExporter(ctx, insecure)
		}

		if err != nil {
			return nil, err
		}

		options = append(options, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	}

	options = append(options, sdkmetric.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attribute.ToKeyValues(resourceAttrs)...)))
	provider := sdkmetric.NewMeterProvider(options...)
	meter := provider.Meter(serviceName)

	if err := initMetricFields(meter, metricsStruct); err != nil {
		return nil, err
	}

	return provider.Shutdown, nil
}
