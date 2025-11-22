package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinybluerobots/gotel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// setupTestTracer creates a tracer with an in-memory exporter for testing
func setupTestTracer(t *testing.T) *tracetest.InMemoryExporter {
	exporter := tracetest.NewInMemoryExporter()
	resourceAttrs := attribute.ResourceAttributes("test-service", "1.0.0", "test", "testhost")
	_, err := InitTracing(
		t.Context(),
		"test-service",
		resourceAttrs,
		sdktrace.WithSyncer(exporter),
	)
	require.NoError(t, err)

	return exporter
}

func TestNewSpan(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span", attribute.New("key", "value"))
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1, "expected 1 span")
	assert.Equal(t, "test-span", spans[0].Name)
}

func TestSpan_AddEvent(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span")
	span.AddEvent("test-event", attribute.New("event-key", "event-value"))
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Len(t, spans[0].Events, 1, "expected 1 event")
	assert.Equal(t, "test-event", spans[0].Events[0].Name)
}

func TestSpan_SetAttributes(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span")
	span.SetAttributes(
		attribute.New("string", "value"),
		attribute.New("int", 42),
	)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.GreaterOrEqual(t, len(spans[0].Attributes), 2, "expected at least 2 attributes")
}

func TestSpan_RecordError(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span")
	testErr := assert.AnError
	span.RecordError(testErr)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	require.Len(t, spans[0].Events, 1, "expected error event")
	assert.Equal(t, "exception", spans[0].Events[0].Name)
}

func TestSpan_RecordErrorAndSetStatus(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span")
	testErr := assert.AnError
	span.RecordErrorAndSetStatus(testErr)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "Error", spans[0].Status.Code.String())
}

func TestSpan_SetStatus(t *testing.T) {
	tests := []struct {
		name                string
		code                StatusCode
		description         string
		expectedCode        string
		expectedDescription string
	}{
		// Note: OTel spec ignores description for Ok status
		{"Ok status", StatusOk, "operation completed", "Ok", ""},
		{"Error status", StatusError, "something failed", "Error", "something failed"},
		{"Unset status", StatusUnset, "", "Unset", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := setupTestTracer(t)
			ctx := t.Context()

			_, span := NewSpan(ctx, "test-span")
			span.SetStatus(tt.code, tt.description)
			span.End()

			spans := exporter.GetSpans()
			require.Len(t, spans, 1)
			assert.Equal(t, tt.expectedCode, spans[0].Status.Code.String())
			assert.Equal(t, tt.expectedDescription, spans[0].Status.Description)
		})
	}
}

func TestSpan_SetOk(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span")
	span.SetOk()
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "Ok", spans[0].Status.Code.String())
}

func TestStartChildSpan(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	ctx, parentSpan := NewSpan(ctx, "parent-span")
	carrier := TraceHeaders(ctx)
	_, childSpan := NewChildSpan(t.Context(), carrier, "child-span")
	childSpan.End()
	parentSpan.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 2, "expected 2 spans")

	// Find parent and child
	var parent, child *tracetest.SpanStub

	for i := range spans {
		switch spans[i].Name {
		case "parent-span":
			parent = &spans[i]
		case "child-span":
			child = &spans[i]
		}
	}

	require.NotNil(t, parent, "parent span not found")
	require.NotNil(t, child, "child span not found")
	assert.Equal(t, parent.SpanContext.SpanID(), child.Parent.SpanID(), "child should reference parent")
}

func TestTraceHeaders(t *testing.T) {
	setupTestTracer(t)

	ctx := t.Context()

	ctx, span := NewSpan(ctx, "test-span")
	defer span.End()

	headers := TraceHeaders(ctx)
	assert.Contains(t, headers, "traceparent", "expected traceparent header")
}

func TestSpanAttributes(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	_, span := NewSpan(ctx, "test-span",
		attribute.New("string", "value"),
		attribute.New("int", 42),
		attribute.New("bool", true),
		attribute.New("float", 3.14),
	)
	span.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.GreaterOrEqual(t, len(spans[0].Attributes), 4, "expected at least 4 attributes")
}

func TestMultipleSpans(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	ctx1, span1 := NewSpan(ctx, "span-1")
	span1.End()

	ctx2, span2 := NewSpan(ctx, "span-2")
	span2.End()

	_ = ctx1
	_ = ctx2

	spans := exporter.GetSpans()
	require.Len(t, spans, 2, "expected 2 spans")
}

func TestNestedChildSpans(t *testing.T) {
	exporter := setupTestTracer(t)
	ctx := t.Context()

	ctx, parentSpan := NewSpan(ctx, "parent")
	childCtx, child := NewChildSpan(t.Context(), TraceHeaders(ctx), "child-1")
	_, grandchild := NewChildSpan(t.Context(), TraceHeaders(childCtx), "grandchild")
	grandchild.End()
	child.End()
	parentSpan.End()

	spans := exporter.GetSpans()
	require.Len(t, spans, 3, "expected 3 spans")
}
