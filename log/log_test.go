package log

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tinybluerobots/gotel/attribute"
)

// captureOutput captures log output during test using the public InitLogger
func captureOutput(t *testing.T, logLevel string) *bytes.Buffer {
	buf := &bytes.Buffer{}

	resourceAttrs := attribute.ResourceAttributes("test-service", "1.0.0", "test", "testhost")
	handler, err := NewJSONHandler(buf, resourceAttrs, logLevel)
	require.NoError(t, err)

	_, err = InitLogger(
		t.Context(),
		resourceAttrs,
		handler,
	)
	require.NoError(t, err)

	return buf
}

func TestInfo(t *testing.T) {
	buf := captureOutput(t, "INFO")
	ctx := t.Context()

	Info(ctx, "test message", attribute.New("key", "value"))

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestDebug(t *testing.T) {
	buf := captureOutput(t, "DEBUG")
	ctx := t.Context()

	Debug(ctx, "debug message", attribute.New("debug-key", "debug-value"))

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "debug message", logEntry["msg"])
	assert.Equal(t, "DEBUG", logEntry["level"])
	assert.Equal(t, "debug-value", logEntry["debug-key"])
}

func TestWarn(t *testing.T) {
	buf := captureOutput(t, "WARN")
	ctx := t.Context()

	Warn(ctx, "warning message", attribute.New("warn-key", "warn-value"))

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "warning message", logEntry["msg"])
	assert.Equal(t, "WARN", logEntry["level"])
	assert.Equal(t, "warn-value", logEntry["warn-key"])
}

func TestError(t *testing.T) {
	buf := captureOutput(t, "ERROR")
	ctx := t.Context()

	Error(ctx, assert.AnError, attribute.New("error-key", "error-value"))

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, assert.AnError.Error(), logEntry["msg"])
	assert.Equal(t, "ERROR", logEntry["level"])
	assert.Equal(t, "error-value", logEntry["error-key"])
}

func TestMultipleAttributes(t *testing.T) {
	buf := captureOutput(t, "INFO")
	ctx := t.Context()

	Info(ctx, "test message",
		attribute.New("string", "value"),
		attribute.New("int", 42),
		attribute.New("bool", true),
		attribute.New("float", 3.14),
	)

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "value", logEntry["string"])
	assert.InDelta(t, 42, logEntry["int"], 0.001)
	assert.Equal(t, true, logEntry["bool"])
	assert.InDelta(t, 3.14, logEntry["float"], 0.001)
}

func TestLogLevelFiltering(t *testing.T) {
	// Set level to WARN - DEBUG and INFO should be filtered
	buf := captureOutput(t, "WARN")
	ctx := t.Context()

	Debug(ctx, "debug message")
	Info(ctx, "info message")

	// Buffer should be empty since both are below WARN level
	assert.Empty(t, buf.String(), "expected no output for logs below WARN level")
}

func TestNoAttributes(t *testing.T) {
	buf := captureOutput(t, "INFO")
	ctx := t.Context()

	Info(ctx, "message without attributes")

	var logEntry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "message without attributes", logEntry["msg"])
}
