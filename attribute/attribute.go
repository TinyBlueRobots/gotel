// Package attribute provides a simplified wrapper around OpenTelemetry attributes.
// It offers automatic type detection and standard resource attribute creation.
package attribute

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

// Attr wraps an OpenTelemetry KeyValue attribute.
type Attr struct {
	attribute.KeyValue
}

func new[T any](key string, value T, convert func(string, T) attribute.KeyValue) Attr {
	return Attr{KeyValue: convert(key, value)}
}

// New creates an attribute with automatic type detection.
// Supported types: bool, []bool, float64, []float64, int, []int, int64, []int64, string, []string.
// Other types are converted using fmt.Stringer or formatted with %v.
func New(key string, value any) Attr {
	switch v := value.(type) {
	case bool:
		return new(key, v, attribute.Bool)
	case []bool:
		return new(key, v, attribute.BoolSlice)
	case float64:
		return new(key, v, attribute.Float64)
	case []float64:
		return new(key, v, attribute.Float64Slice)
	case int:
		return new(key, v, attribute.Int)
	case []int:
		return new(key, v, attribute.IntSlice)
	case int64:
		return new(key, v, attribute.Int64)
	case []int64:
		return new(key, v, attribute.Int64Slice)
	case string:
		return new(key, v, attribute.String)
	case []string:
		return new(key, v, attribute.StringSlice)
	case fmt.Stringer:
		return new(key, v.String(), attribute.String)
	default:
		return Attr{KeyValue: attribute.String(key, fmt.Sprintf("%v", v))}
	}
}

// ResourceAttributes creates standard resource attributes for a service.
func ResourceAttributes(serviceName string, serviceVersion string, environment string, hostname string) []Attr {
	return []Attr{
		{semconv.DeploymentEnvironmentNameKey.String(environment)},
		{semconv.HostNameKey.String(hostname)},
		{semconv.ProcessExecutableNameKey.String(serviceName)},
		{semconv.ServiceNameKey.String(serviceName)},
		{semconv.ServiceVersionKey.String(serviceVersion)},
	}
}

// ToKeyValues converts a slice of Attr to an OpenTelemetry KeyValue slice.
func ToKeyValues(attrs []Attr) []attribute.KeyValue {
	result := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		result[i] = attr.KeyValue
	}

	return result
}
