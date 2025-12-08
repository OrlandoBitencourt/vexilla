package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTracer(t *testing.T) {
	tracer := NewTracer("test-tracer")
	require.NotNil(t, tracer)
	assert.NotNil(t, tracer.Tracer())
}

func TestTracer_Tracer(t *testing.T) {
	tracer := NewTracer("vexilla")
	underlying := tracer.Tracer()

	require.NotNil(t, underlying)
	// Verify we can start spans
	_, span := underlying.Start(context.Background(), "test-span")
	assert.NotNil(t, span)
	span.End()
}
