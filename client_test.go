package vexilla

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_StartStop tests client lifecycle
func TestClient_StartStop(t *testing.T) {
	// Create mock Flagr server
	server := NewMockFlagrServer(t)
	defer server.Close()

	// Add test flags
	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Distributions:  []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:         1,
				Key:        "enabled",
				Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`true`)},
			},
		},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithRefreshInterval(10*time.Minute),
	)
	require.NoError(t, err)

	// Test Start
	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err, "Start should succeed")

	// Give cache time to initialize
	time.Sleep(100 * time.Millisecond)

	// Test Stop
	err = client.Stop()
	assert.NoError(t, err, "Stop should succeed")
}

// TestClient_Bool tests boolean flag evaluation
func TestClient_Bool(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "bool-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "country",
						Operator: domain.OperatorEQ,
						Value:    "BR",
					},
				},
				Distributions: []domain.Distribution{
					{VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "enabled",
				Attachment: map[string]json.RawMessage{
					"value": json.RawMessage(`true`),
				},
			},
		},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithOnlyEnabled(true),
	)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	require.NoError(t, client.Sync(ctx))

	evalCtx := NewContext("user-123").WithAttribute("country", "BR")
	resp, err := client.Evaluate(ctx, "bool-flag", evalCtx)

	require.NoError(t, err)
	require.True(t, resp.IsEnabled())
}

// TestClient_String tests string flag evaluation
func TestClient_String(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "theme-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           1,
				RolloutPercent: 100,
				Constraints:    []domain.Constraint{},
				Distributions: []domain.Distribution{
					{
						ID:        1,
						VariantID: 1,
						Percent:   100,
					},
				},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "dark",
				Attachment: map[string]json.RawMessage{
					"value": json.RawMessage(`"dark-theme"`),
				},
			},
		},
	})

	client, err := New(WithFlagrEndpoint(server.URL), WithOnlyEnabled(true))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	require.NoError(t, client.Sync(ctx))

	evalCtx := NewContext("user-456")
	theme := client.String(ctx, "theme-flag", evalCtx, "light")

	assert.Equal(t, "dark-theme", theme)
}

// TestClient_Int tests integer flag evaluation
func TestClient_Int(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      3,
		Key:     "limit-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Distributions:  []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:         1,
				Key:        "premium",
				Attachment: map[string]json.RawMessage{"value": json.RawMessage(`1000`)},
			},
		},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("user-789")
	limit := client.Int(ctx, "limit-flag", evalCtx, 100)

	assert.Equal(t, 1000, limit)
}

// TestClient_Evaluate tests detailed evaluation
func TestClient_Evaluate(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      4,
		Key:     "detailed-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Distributions:  []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "variant-a",
				Attachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
					"color":   json.RawMessage(`"blue"`),
				},
			},
		},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("user-abc")
	result, err := client.Evaluate(ctx, "detailed-flag", evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "detailed-flag", result.FlagKey)
	assert.Equal(t, "variant-a", result.VariantKey)
	assert.True(t, result.IsEnabled())
	assert.Equal(t, "blue", result.GetString("color", ""))
}

// TestClient_InvalidateFlag tests flag invalidation
func TestClient_InvalidateFlag(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      5,
		Key:     "invalidate-me",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	// Invalidate flag
	err = client.InvalidateFlag(ctx, "invalidate-me")
	assert.NoError(t, err)
}

// TestClient_InvalidateAll tests clearing all flags
func TestClient_InvalidateAll(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{ID: 1, Key: "flag1", Enabled: true})
	server.AddFlag(domain.Flag{ID: 2, Key: "flag2", Enabled: true})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	err = client.InvalidateAll(ctx)
	assert.NoError(t, err)
}

// TestClient_Metrics tests metrics retrieval
func TestClient_Metrics(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      6,
		Key:     "metrics-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	// Make some evaluations
	evalCtx := NewContext("user-metrics")
	client.Bool(ctx, "metrics-flag", evalCtx)
	client.Bool(ctx, "metrics-flag", evalCtx)

	metrics := client.Metrics()

	assert.NotZero(t, metrics.Storage.KeysAdded, "Should have added keys")
	assert.False(t, metrics.LastRefresh.IsZero(), "Should have refreshed")
	assert.False(t, metrics.CircuitOpen, "Circuit should be closed")
}

// TestClient_MissingFlag tests fallback behavior
func TestClient_MissingFlag(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithFallbackStrategy("fail_closed"),
	)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("user-missing")
	enabled := client.Bool(ctx, "nonexistent-flag", evalCtx)

	assert.False(t, enabled, "Missing flag with fail_closed should return false")
}

// TestClient_RemoteEvaluation tests remote strategy
func TestClient_RemoteEvaluation(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	// Flag with partial rollout (requires remote evaluation)
	server.AddFlag(domain.Flag{
		ID:      7,
		Key:     "gradual-rollout",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 50, // Partial rollout = remote evaluation
				Distributions:  []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "enabled", Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`true`)}},
		},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("user-remote")
	result, err := client.Evaluate(ctx, "gradual-rollout", evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "gradual-rollout", result.FlagKey)
}

// TestClient_ConcurrentAccess tests thread safety
func TestClient_ConcurrentAccess(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      8,
		Key:     "concurrent-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	// Concurrent evaluations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			evalCtx := NewContext(fmt.Sprintf("user-%d", id))
			client.Bool(ctx, "concurrent-flag", evalCtx)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestClient_Start_WithWebhookServer tests starting with webhook server
func TestClient_Start_WithWebhookServer(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithWebhookInvalidation(WebhookConfig{
			Port:   28001,
			Secret: "test-secret",
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = client.Stop()
	assert.NoError(t, err)
}

// TestClient_Start_WithAdminServer tests starting with admin server
func TestClient_Start_WithAdminServer(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithAdminServer(AdminConfig{
			Port: 29001,
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = client.Stop()
	assert.NoError(t, err)
}

// TestClient_Sync_NilCache tests Sync with nil cache
func TestClient_Sync_NilCache(t *testing.T) {
	client := &Client{
		cache: nil,
	}

	err := client.Sync(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache not initialized")
}

// TestClient_Evaluate_NotFound tests evaluating a non-existent flag
func TestClient_Evaluate_NotFound(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	client, err := New(WithFlagrEndpoint(server.URL))
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("user-test")
	result, err := client.Evaluate(ctx, "nonexistent", evalCtx)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// BenchmarkClient_Bool benchmarks boolean evaluation
func BenchmarkClient_Bool(b *testing.B) {
	server := NewMockFlagrServer(&testing.T{})
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      99,
		Key:     "bench-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "on", Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`true`)}},
		},
	})

	client, _ := New(WithFlagrEndpoint(server.URL))
	ctx := context.Background()
	client.Start(ctx)
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	evalCtx := NewContext("bench-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Bool(ctx, "bench-flag", evalCtx)
	}
}
