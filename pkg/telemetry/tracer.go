package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer creates a new tracer
func NewTracer(tracerName string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(tracerName),
	}
}

// Tracer returns the underlying tracer
func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}
