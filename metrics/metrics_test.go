package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinybluerobots/gotel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// TestMetrics is a sample metrics struct for testing
type TestMetrics struct {
	Counter                *Int64Counter
	FloatCounter           *Float64Counter
	FloatGauge             *Float64Gauge
	FloatHistogram         *Float64Histogram
	FloatUpDown            *Float64UpDownCounter
	Gauge                  *Int64Gauge
	Histogram              *Int64Histogram
	UpDown                 *Int64UpDownCounter
	ObservableCounter      *Int64ObservableCounter
	ObservableFloatCounter *Float64ObservableCounter
	ObservableUpDown       *Int64ObservableUpDownCounter
	ObservableFloatUpDown  *Float64ObservableUpDownCounter
	ObservableGauge        *Int64ObservableGauge
	ObservableFloatGauge   *Float64ObservableGauge
}

// initTestMetrics initializes TestMetrics with a test reader
func initTestMetrics(t *testing.T) (*TestMetrics, *sdkmetric.ManualReader) {
	reader := sdkmetric.NewManualReader()
	m := &TestMetrics{}
	resourceAttrs := attribute.ResourceAttributes("test-service", "1.0.0", "test", "testhost")
	_, err := InitMetrics(
		t.Context(),
		"test-service",
		resourceAttrs,
		m,
		sdkmetric.WithReader(reader),
	)
	require.NoError(t, err)

	return m, reader
}

// findMetric searches for a metric by name in ResourceMetrics
func findMetric(rm metricdata.ResourceMetrics, name string) *metricdata.Metrics {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return &m
			}
		}
	}

	return nil
}

func TestMetrics_Retrieval(t *testing.T) {
	m, _ := initTestMetrics(t)

	retrieved := Metrics[TestMetrics]()
	require.NotNil(t, retrieved, "Metrics() returned nil")
	assert.Equal(t, m, retrieved, "Metrics() returned different instance")
}

func TestInt64Counter_Add(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.Counter.Add(ctx, 5, attribute.New("key", "value"))
	m.Counter.Add(ctx, 3)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "Counter")
	require.NotNil(t, metric, "Counter metric not found")

	sum, ok := metric.Data.(metricdata.Sum[int64])
	require.True(t, ok, "expected Sum[int64], got %T", metric.Data)
	assert.NotEmpty(t, sum.DataPoints, "no data points recorded")
}

func TestFloat64Counter_Add(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.FloatCounter.Add(ctx, 5.5, attribute.New("key", "value"))
	m.FloatCounter.Add(ctx, 3.3)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "FloatCounter")
	require.NotNil(t, metric, "FloatCounter metric not found")

	sum, ok := metric.Data.(metricdata.Sum[float64])
	require.True(t, ok, "expected Sum[float64], got %T", metric.Data)
	assert.NotEmpty(t, sum.DataPoints, "no data points recorded")
}

func TestInt64UpDownCounter_Add(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.UpDown.Add(ctx, 10)
	m.UpDown.Add(ctx, -3)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "UpDown")
	require.NotNil(t, metric, "UpDown metric not found")
}

func TestFloat64UpDownCounter_Add(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.FloatUpDown.Add(ctx, 10.5)
	m.FloatUpDown.Add(ctx, -3.2)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "FloatUpDown")
	require.NotNil(t, metric, "FloatUpDown metric not found")
}

func TestInt64Gauge_Record(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.Gauge.Record(ctx, 42, attribute.New("instance", "test"))

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "Gauge")
	require.NotNil(t, metric, "Gauge metric not found")
}

func TestFloat64Gauge_Record(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.FloatGauge.Record(ctx, 42.5, attribute.New("instance", "test"))

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "FloatGauge")
	require.NotNil(t, metric, "FloatGauge metric not found")
}

func TestInt64Histogram_Record(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.Histogram.Record(ctx, 100, attribute.New("endpoint", "/api"))
	m.Histogram.Record(ctx, 150)
	m.Histogram.Record(ctx, 200)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "Histogram")
	require.NotNil(t, metric, "Histogram metric not found")

	hist, ok := metric.Data.(metricdata.Histogram[int64])
	require.True(t, ok, "expected Histogram[int64], got %T", metric.Data)
	assert.NotEmpty(t, hist.DataPoints, "no data points recorded")
}

func TestFloat64Histogram_Record(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	m.FloatHistogram.Record(ctx, 100.5, attribute.New("endpoint", "/api"))
	m.FloatHistogram.Record(ctx, 150.5)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	metric := findMetric(rm, "FloatHistogram")
	require.NotNil(t, metric, "FloatHistogram metric not found")

	hist, ok := metric.Data.(metricdata.Histogram[float64])
	require.True(t, ok, "expected Histogram[float64], got %T", metric.Data)
	assert.NotEmpty(t, hist.DataPoints, "no data points recorded")
}

// Test nil receiver safety - methods should not panic on nil
func TestNilReceiverSafety(t *testing.T) {
	ctx := t.Context()

	t.Run("Int64Counter", func(t *testing.T) {
		var c *Int64Counter

		assert.NotPanics(t, func() { c.Add(ctx, 1) })
	})

	t.Run("Float64Counter", func(t *testing.T) {
		var c *Float64Counter

		assert.NotPanics(t, func() { c.Add(ctx, 1.0) })
	})

	t.Run("Int64UpDownCounter", func(t *testing.T) {
		var c *Int64UpDownCounter

		assert.NotPanics(t, func() { c.Add(ctx, 1) })
	})

	t.Run("Float64UpDownCounter", func(t *testing.T) {
		var c *Float64UpDownCounter

		assert.NotPanics(t, func() { c.Add(ctx, 1.0) })
	})

	t.Run("Int64Gauge", func(t *testing.T) {
		var g *Int64Gauge

		assert.NotPanics(t, func() { g.Record(ctx, 1) })
	})

	t.Run("Float64Gauge", func(t *testing.T) {
		var g *Float64Gauge

		assert.NotPanics(t, func() { g.Record(ctx, 1.0) })
	})

	t.Run("Int64Histogram", func(t *testing.T) {
		var h *Int64Histogram

		assert.NotPanics(t, func() { h.Record(ctx, 1) })
	})

	t.Run("Float64Histogram", func(t *testing.T) {
		var h *Float64Histogram

		assert.NotPanics(t, func() { h.Record(ctx, 1.0) })
	})
}

func TestAttributes(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	// Record with multiple attributes
	m.Counter.Add(ctx, 1,
		attribute.New("string", "value"),
		attribute.New("int", 42),
		attribute.New("bool", true),
		attribute.New("float", 3.14),
	)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "Counter")
	require.NotNil(t, foundMetric, "Counter metric not found")

	sum := foundMetric.Data.(metricdata.Sum[int64])
	require.NotEmpty(t, sum.DataPoints, "no data points recorded")

	dp := sum.DataPoints[0]
	assert.Equal(t, 4, dp.Attributes.Len(), "expected 4 attributes")
}

func TestInt64ObservableCounter_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableCounter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveInt64(m.ObservableCounter.int64ObservableCounter, 42)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableCounter")
	require.NotNil(t, foundMetric, "ObservableCounter metric not found")

	sum, ok := foundMetric.Data.(metricdata.Sum[int64])
	require.True(t, ok, "expected Sum[int64], got %T", foundMetric.Data)
	require.NotEmpty(t, sum.DataPoints, "no data points recorded")
	assert.Equal(t, int64(42), sum.DataPoints[0].Value)
}

func TestFloat64ObservableCounter_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableFloatCounter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveFloat64(m.ObservableFloatCounter.float64ObservableCounter, 42.5)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableFloatCounter")
	require.NotNil(t, foundMetric, "ObservableFloatCounter metric not found")

	sum, ok := foundMetric.Data.(metricdata.Sum[float64])
	require.True(t, ok, "expected Sum[float64], got %T", foundMetric.Data)
	require.NotEmpty(t, sum.DataPoints, "no data points recorded")
	assert.InDelta(t, 42.5, sum.DataPoints[0].Value, 0.001)
}

func TestInt64ObservableGauge_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableGauge.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveInt64(m.ObservableGauge.int64ObservableGauge, 100)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableGauge")
	require.NotNil(t, foundMetric, "ObservableGauge metric not found")

	gauge, ok := foundMetric.Data.(metricdata.Gauge[int64])
	require.True(t, ok, "expected Gauge[int64], got %T", foundMetric.Data)
	require.NotEmpty(t, gauge.DataPoints, "no data points recorded")
	assert.Equal(t, int64(100), gauge.DataPoints[0].Value)
}

func TestFloat64ObservableGauge_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableFloatGauge.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveFloat64(m.ObservableFloatGauge.float64ObservableGauge, 99.9)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableFloatGauge")
	require.NotNil(t, foundMetric, "ObservableFloatGauge metric not found")

	gauge, ok := foundMetric.Data.(metricdata.Gauge[float64])
	require.True(t, ok, "expected Gauge[float64], got %T", foundMetric.Data)
	require.NotEmpty(t, gauge.DataPoints, "no data points recorded")
	assert.InDelta(t, 99.9, gauge.DataPoints[0].Value, 0.001)
}

func TestInt64ObservableUpDownCounter_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableUpDown.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveInt64(m.ObservableUpDown.int64ObservableUpDownCounter, -5)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableUpDown")
	require.NotNil(t, foundMetric, "ObservableUpDown metric not found")

	sum, ok := foundMetric.Data.(metricdata.Sum[int64])
	require.True(t, ok, "expected Sum[int64], got %T", foundMetric.Data)
	require.NotEmpty(t, sum.DataPoints, "no data points recorded")
	assert.Equal(t, int64(-5), sum.DataPoints[0].Value)
}

func TestFloat64ObservableUpDownCounter_RegisterCallback(t *testing.T) {
	m, reader := initTestMetrics(t)
	ctx := t.Context()

	_, err := m.ObservableFloatUpDown.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		o.ObserveFloat64(m.ObservableFloatUpDown.float64ObservableUpDownCounter, -3.5)
		return nil
	})
	require.NoError(t, err)

	rm := metricdata.ResourceMetrics{}
	require.NoError(t, reader.Collect(ctx, &rm))

	foundMetric := findMetric(rm, "ObservableFloatUpDown")
	require.NotNil(t, foundMetric, "ObservableFloatUpDown metric not found")

	sum, ok := foundMetric.Data.(metricdata.Sum[float64])
	require.True(t, ok, "expected Sum[float64], got %T", foundMetric.Data)
	require.NotEmpty(t, sum.DataPoints, "no data points recorded")
	assert.InDelta(t, -3.5, sum.DataPoints[0].Value, 0.001)
}
