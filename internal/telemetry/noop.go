// pkg/telemetry/noop.go
package telemetry

import (
	"context"
	"time"
)

// NoOpProvider is a telemetry provider that does nothing
// Useful for testing or when telemetry is disabled
type NoOpProvider struct{}

// NewNoOp creates a new no-op telemetry provider
func NewNoOp() *NoOpProvider {
	return &NoOpProvider{}
}

// StartSpan creates a no-op span
func (n *NoOpProvider) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	return ctx, &NoOpSpan{}
}

// RecordCacheHit does nothing
func (n *NoOpProvider) RecordCacheHit(ctx context.Context, flagKey string) {}

// RecordCacheMiss does nothing
func (n *NoOpProvider) RecordCacheMiss(ctx context.Context, flagKey string) {}

// RecordEvaluation does nothing
func (n *NoOpProvider) RecordEvaluation(ctx context.Context, flagKey string, strategy string, duration time.Duration) {
}

// RecordRefresh does nothing
func (n *NoOpProvider) RecordRefresh(ctx context.Context, success bool, duration time.Duration, flagCount int) {
}

// RecordCircuitState does nothing
func (n *NoOpProvider) RecordCircuitState(ctx context.Context, state string) {}

// Shutdown does nothing
func (n *NoOpProvider) Shutdown(ctx context.Context) error {
	return nil
}

// NoOpSpan is a span that does nothing
type NoOpSpan struct{}

// End does nothing
func (n *NoOpSpan) End() {}

// SetAttributes does nothing
func (n *NoOpSpan) SetAttributes(attrs ...Attribute) {}

// RecordError does nothing
func (n *NoOpSpan) RecordError(err error) {}

// AddEvent does nothing
func (n *NoOpSpan) AddEvent(name string, attrs ...Attribute) {}
