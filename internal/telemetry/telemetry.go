package telemetry

import (
	"context"
	"time"
)

// Provider defines the interface for telemetry providers
type Provider interface {
	// Tracer operations
	StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span)

	// Metrics operations
	RecordCacheHit(ctx context.Context, flagKey string)
	RecordCacheMiss(ctx context.Context, flagKey string)
	RecordEvaluation(ctx context.Context, flagKey string, strategy string, duration time.Duration)
	RecordRefresh(ctx context.Context, success bool, duration time.Duration, flagCount int)
	RecordCircuitState(ctx context.Context, state string)

	// Lifecycle
	Shutdown(ctx context.Context) error
}

// Span represents a trace span
type Span interface {
	// End completes the span
	End()

	// SetAttributes sets attributes on the span
	SetAttributes(attrs ...Attribute)

	// RecordError records an error
	RecordError(err error)

	// AddEvent adds an event to the span
	AddEvent(name string, attrs ...Attribute)
}

// SpanOption configures span creation
type SpanOption func(*SpanConfig)

// SpanConfig holds span configuration
type SpanConfig struct {
	Attributes []Attribute
}

// Attribute represents a key-value attribute
type Attribute struct {
	Key   string
	Value interface{}
}

// WithAttributes adds attributes to a span
func WithAttributes(attrs ...Attribute) SpanOption {
	return func(c *SpanConfig) {
		c.Attributes = append(c.Attributes, attrs...)
	}
}

// String creates a string attribute
func String(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int creates an int attribute
func Int(key string, value int) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int64 creates an int64 attribute
func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Bool creates a bool attribute
func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

// Float64 creates a float64 attribute
func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Duration creates a duration attribute
func Duration(key string, value time.Duration) Attribute {
	return Attribute{Key: key, Value: value.Milliseconds()}
}
