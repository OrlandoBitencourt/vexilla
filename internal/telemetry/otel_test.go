package telemetry

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelTest initializes OpenTelemetry for testing
func setupOTelTest(t *testing.T) (*OTelProvider, func()) {
	t.Helper()

	// Set up trace provider
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	// Set up metric provider
	mp := metric.NewMeterProvider()
	otel.SetMeterProvider(mp)

	provider, err := NewOTel()
	if err != nil {
		t.Fatalf("failed to create OTel provider: %v", err)
	}

	cleanup := func() {
		ctx := context.Background()
		_ = provider.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}

	return provider, cleanup
}

func TestNewOTel(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	if provider.tracer == nil {
		t.Error("expected non-nil tracer")
	}
	if provider.meter == nil {
		t.Error("expected non-nil meter")
	}
}

func TestOTelProvider_InitMetrics(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	if provider.cacheHits == nil {
		t.Error("expected cacheHits to be initialized")
	}
	if provider.cacheMisses == nil {
		t.Error("expected cacheMisses to be initialized")
	}
	if provider.evaluations == nil {
		t.Error("expected evaluations to be initialized")
	}
	if provider.refreshDuration == nil {
		t.Error("expected refreshDuration to be initialized")
	}
	if provider.refreshSuccess == nil {
		t.Error("expected refreshSuccess to be initialized")
	}
	if provider.refreshFailure == nil {
		t.Error("expected refreshFailure to be initialized")
	}
	if provider.circuitState == nil {
		t.Error("expected circuitState to be initialized")
	}
}

func TestOTelProvider_GetCircuitStateValue(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	tests := []struct {
		state    string
		expected int64
	}{
		{"closed", 0},
		{"open", 1},
		{"half-open", 2},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			provider.currentCircuitState = tt.state
			got := provider.getCircuitStateValue()
			if got != tt.expected {
				t.Errorf("getCircuitStateValue(%q) = %d, want %d", tt.state, got, tt.expected)
			}
		})
	}
}

func TestOTelProvider_StartSpan(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	newCtx, span := provider.StartSpan(ctx, "test-span")

	if newCtx == ctx {
		t.Error("expected new context")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}

	otelSpan, ok := span.(*OTelSpan)
	if !ok {
		t.Errorf("expected *OTelSpan, got %T", span)
	}
	if otelSpan.span == nil {
		t.Error("expected non-nil underlying span")
	}
	if otelSpan.provider != provider {
		t.Error("expected span to reference provider")
	}
}

func TestOTelProvider_StartSpanWithAttributes(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	attrs := []SpanOption{
		WithAttributes(
			String("string_key", "string_value"),
			Int("int_key", 42),
			Int64("int64_key", int64(123)),
			Bool("bool_key", true),
			Float64("float_key", 3.14),
		),
	}

	newCtx, span := provider.StartSpan(ctx, "test-span", attrs...)

	if newCtx == ctx {
		t.Error("expected new context")
	}
	if span == nil {
		t.Fatal("expected non-nil span")
	}
}

func TestOTelProvider_ConvertAttribute(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	tests := []struct {
		name     string
		attr     Attribute
		wantType string
	}{
		{"string", String("key", "value"), "STRING"},
		{"int", Int("key", 42), "INT64"},
		{"int64", Int64("key", int64(123)), "INT64"},
		{"bool", Bool("key", true), "BOOL"},
		{"float64", Float64("key", 3.14), "FLOAT64"},
		{"unknown", Attribute{Key: "key", Value: struct{}{}}, "STRING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.convertAttribute(tt.attr)
			if string(result.Key) != tt.attr.Key {
				t.Errorf("key mismatch: got %s, want %s", result.Key, tt.attr.Key)
			}
		})
	}
}

func TestOTelProvider_RecordCacheHit(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	// Should not panic
	provider.RecordCacheHit(ctx, "test-flag")
}

func TestOTelProvider_RecordCacheMiss(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	// Should not panic
	provider.RecordCacheMiss(ctx, "test-flag")
}

func TestOTelProvider_RecordEvaluation(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	// Should not panic
	provider.RecordEvaluation(ctx, "test-flag", "gradual", 10*time.Millisecond)
}

func TestOTelProvider_RecordRefresh(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("successful refresh", func(t *testing.T) {
		provider.RecordRefresh(ctx, true, 100*time.Millisecond, 5)
	})

	t.Run("failed refresh", func(t *testing.T) {
		provider.RecordRefresh(ctx, false, 200*time.Millisecond, 0)
	})
}

func TestOTelProvider_RecordCircuitState(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()

	states := []string{"closed", "open", "half-open"}
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			provider.RecordCircuitState(ctx, state)
			if provider.currentCircuitState != state {
				t.Errorf("expected state %s, got %s", state, provider.currentCircuitState)
			}
		})
	}
}

func TestOTelProvider_Shutdown(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	err := provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOTelSpan_End(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-span")

	// Should not panic
	span.End()
}

func TestOTelSpan_SetAttributes(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-span")

	// Should not panic
	span.SetAttributes(
		String("key1", "value1"),
		Int("key2", 42),
		Bool("key3", true),
	)
}

func TestOTelSpan_RecordError(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-span")

	// Should not panic
	span.RecordError(errors.New("test error"))
}

func TestOTelSpan_AddEvent(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	_, span := provider.StartSpan(ctx, "test-span")

	// Should not panic
	span.AddEvent("event1")
	span.AddEvent("event2", String("key", "value"))
}

func TestOTelProvider_ImplementsInterface(t *testing.T) {
	var _ Provider = (*OTelProvider)(nil)
}

func TestOTelSpan_ImplementsInterface(t *testing.T) {
	var _ Span = (*OTelSpan)(nil)
}

func TestOTelProvider_ConcurrentUsage(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()
	done := make(chan bool)

	// Run multiple operations concurrently
	for i := 0; i < 10; i++ {
		go func() {
			provider.RecordCacheHit(ctx, "flag")
			provider.RecordCacheMiss(ctx, "flag")
			provider.RecordEvaluation(ctx, "flag", "strategy", time.Millisecond)
			provider.RecordRefresh(ctx, true, time.Millisecond, 1)
			provider.RecordCircuitState(ctx, "closed")

			_, span := provider.StartSpan(ctx, "test")
			span.SetAttributes(String("k", "v"))
			span.RecordError(errors.New("err"))
			span.AddEvent("event")
			span.End()

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestOTelSpan_NestedSpans(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create parent span
	ctx1, span1 := provider.StartSpan(ctx, "parent-span")
	defer span1.End()

	// Create child span
	ctx2, span2 := provider.StartSpan(ctx1, "child-span")
	defer span2.End()

	if ctx1 == ctx2 {
		t.Error("expected different contexts for parent and child spans")
	}
}

func TestOTelProvider_MetricsWithVariousFlags(t *testing.T) {
	provider, cleanup := setupOTelTest(t)
	defer cleanup()

	ctx := context.Background()

	flags := []string{"flag1", "flag2", "flag3"}
	for _, flag := range flags {
		provider.RecordCacheHit(ctx, flag)
		provider.RecordCacheMiss(ctx, flag)
		provider.RecordEvaluation(ctx, flag, "gradual", time.Millisecond)
	}
}
