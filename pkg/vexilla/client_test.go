package vexilla

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"

	client, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.memoryStore)
	assert.NotNil(t, client.evaluator)
	assert.NotNil(t, client.strategy)
	assert.NotNil(t, client.breaker)
	assert.NotNil(t, client.metrics)
	assert.NotNil(t, client.tracer)
}

func TestNew_InvalidConfig(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "" // Invalid

	client, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestClient_EvaluateBool(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	// Add a test flag to memory store
	testFlag := Flag{
		ID:      1,
		Key:     "test_flag",
		Enabled: true,
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Constraints: []Constraint{
					{Property: "country", Operator: "EQ", Value: "US"},
				},
				Distributions: []Distribution{
					{VariantKey: "enabled", Percent: 100},
				},
			},
		},
	}
	client.memoryStore.Set(ctx, "test_flag", testFlag)
	client.memoryStore.Wait()

	// Test evaluation
	evalCtx := EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "US",
		},
	}

	result := client.EvaluateBool(ctx, "test_flag", evalCtx)
	assert.True(t, result)
}

func TestClient_EvaluateString(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	testFlag := Flag{
		ID:      1,
		Key:     "theme_flag",
		Enabled: true,
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{VariantKey: "dark", Percent: 100},
				},
			},
		},
	}
	client.memoryStore.Set(ctx, "theme_flag", testFlag)
	client.memoryStore.Wait()

	evalCtx := EvaluationContext{EntityID: "user123"}
	result := client.EvaluateString(ctx, "theme_flag", evalCtx, "light")
	assert.Equal(t, "dark", result)
}

func TestClient_Evaluate_NotFound(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()
	evalCtx := EvaluationContext{EntityID: "user123"}

	result, err := client.Evaluate(ctx, "nonexistent", evalCtx)
	assert.Error(t, err)
	assert.Nil(t, result)

	_, ok := err.(ErrFlagNotFound)
	assert.True(t, ok)
}

func TestClient_InvalidateFlags(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	// Add flags
	testFlag := Flag{ID: 1, Key: "flag1"}
	client.memoryStore.Set(ctx, "flag1", testFlag)
	client.memoryStore.Wait()

	// Verify exists
	_, found := client.memoryStore.Get(ctx, "flag1")
	assert.True(t, found)

	// Invalidate
	err = client.InvalidateFlags(ctx, []string{"flag1"})
	require.NoError(t, err)

	// Verify removed
	_, found = client.memoryStore.Get(ctx, "flag1")
	assert.False(t, found)
}

func TestClient_InvalidateAll(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	// Add multiple flags
	for i := 0; i < 5; i++ {
		flag := Flag{ID: int64(i), Key: fmt.Sprintf("flag%d", i)}
		client.memoryStore.Set(ctx, flag.Key, flag)
	}
	client.memoryStore.Wait()

	// Invalidate all
	err = client.InvalidateAll(ctx)
	require.NoError(t, err)

	// Verify all removed
	for i := 0; i < 5; i++ {
		_, found := client.memoryStore.Get(ctx, fmt.Sprintf("flag%d", i))
		assert.False(t, found)
	}
}

func TestClient_GetStats(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	stats, err := client.GetStats(ctx)
	require.NoError(t, err)

	s, ok := stats.(Stats)
	require.True(t, ok)

	assert.GreaterOrEqual(t, s.KeysAdded, uint64(0))
	assert.False(t, s.CircuitOpen)
}

func TestClient_HealthCheck(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	health, err := client.HealthCheck(ctx)
	require.NoError(t, err)

	h, ok := health.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "healthy", h["status"])
	assert.False(t, h["circuit_open"])
}

func TestClient_Middleware(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	// Create a test handler
	var capturedCtx context.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	wrappedHandler := client.Middleware(handler)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user123")
	rw := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(rw, req)

	// Verify context was injected
	assert.NotNil(t, capturedCtx)

	clientFromCtx, ok := capturedCtx.Value(contextKeyClient).(*Client)
	assert.True(t, ok)
	assert.Equal(t, client, clientFromCtx)

	evalCtx, ok := capturedCtx.Value(contextKeyEvalCtx).(EvaluationContext)
	assert.True(t, ok)
	assert.Equal(t, "user123", evalCtx.EntityID)
}

func TestGetFlagFromContext(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	client, err := New(config)
	require.NoError(t, err)
	defer client.Stop()

	ctx := context.Background()

	// Add test flag
	testFlag := Flag{
		ID:      1,
		Key:     "context_flag",
		Enabled: true,
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{VariantKey: "enabled", Percent: 100},
				},
			},
		},
	}
	client.memoryStore.Set(ctx, "context_flag", testFlag)
	client.memoryStore.Wait()

	// Create context with client and eval context
	ctx = context.WithValue(ctx, contextKeyClient, client)
	ctx = context.WithValue(ctx, contextKeyEvalCtx, EvaluationContext{
		EntityID: "user123",
	})

	// Test GetFlagFromContext
	result, err := GetFlagFromContext(ctx, "context_flag")
	require.NoError(t, err)
	assert.Equal(t, "enabled", result.VariantKey)

	// Test GetFlagBoolFromContext
	enabled := GetFlagBoolFromContext(ctx, "context_flag")
	assert.True(t, enabled)
}
