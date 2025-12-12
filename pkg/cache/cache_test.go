package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/flagr"
	"github.com/OrlandoBitencourt/vexilla/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create RawMessage
func raw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestNew(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)

	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNew_MissingDependencies(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
	}{
		{
			name: "missing flagr client",
			options: []Option{
				WithStorage(storage.NewMockStorage()),
				WithEvaluator(evaluator.New()),
			},
		},
		{
			name: "missing storage",
			options: []Option{
				WithFlagrClient(flagr.NewMockClient()),
				WithEvaluator(evaluator.New()),
			},
		},
		{
			name: "missing evaluator",
			options: []Option{
				WithFlagrClient(flagr.NewMockClient()),
				WithStorage(storage.NewMockStorage()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.options...)
			assert.Error(t, err)
		})
	}
}

func TestCache_Start(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	// Test flag
	testFlag := domain.Flag{
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
				Key:        "on",
				Attachment: map[string]json.RawMessage{"enabled": raw(true)},
			},
		},
	}

	mockFlagr.AddFlag(testFlag)

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)

	mockFlagr.AssertCalled(t, "GetAllFlags", 1)
	mockStorage.AssertCalled(t, "Set", 1)

	c.Stop()
}

func TestCache_Evaluate_LocalStrategy(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	testFlag := domain.Flag{
		ID:      1,
		Key:     "local-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{Property: "country", Operator: domain.OperatorEQ, Value: "BR"},
				},
				Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:         1,
				Key:        "enabled",
				Attachment: map[string]json.RawMessage{"enabled": raw(true)},
			},
		},
	}

	mockStorage.AddFlag(testFlag)

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)
	require.NoError(t, err)

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context:  map[string]interface{}{"country": "BR"},
	}

	result, err := c.Evaluate(ctx, "local-flag", evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "local-flag", result.FlagKey)
	assert.Equal(t, "enabled", result.VariantKey)

	var enabled bool
	json.Unmarshal(result.VariantAttachment["enabled"], &enabled)
	assert.True(t, enabled)

	mockFlagr.AssertCalled(t, "EvaluateFlag", 0)
}

func TestCache_Evaluate_RemoteStrategy(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	testFlag := domain.Flag{
		ID:      1,
		Key:     "remote-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 50,
				Distributions:  []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:         1,
				Key:        "enabled",
				Attachment: map[string]json.RawMessage{"enabled": raw(true)},
			},
		},
	}

	mockStorage.AddFlag(testFlag)

	mockFlagr.EvaluateFlagFunc = func(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
		return &domain.EvaluationResult{
			FlagID:            1,
			FlagKey:           "remote-flag",
			VariantID:         1,
			VariantKey:        "enabled",
			VariantAttachment: map[string]json.RawMessage{"enabled": raw(true)},
		}, nil
	}

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)
	require.NoError(t, err)

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{}

	result, err := c.Evaluate(ctx, "remote-flag", evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "remote-flag", result.FlagKey)

	mockFlagr.AssertCalled(t, "EvaluateFlag", 1)
}

func TestCache_EvaluateBool(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	testFlag := domain.Flag{
		ID:      1,
		Key:     "bool-flag",
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
				Attachment: map[string]json.RawMessage{"enabled": raw(true)},
			},
		},
	}

	mockStorage.AddFlag(testFlag)

	c, _ := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)

	ctx := context.Background()
	result := c.EvaluateBool(ctx, "bool-flag", domain.EvaluationContext{})

	assert.True(t, result)
}

func TestCache_EvaluateString(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	testFlag := domain.Flag{
		ID:      1,
		Key:     "string-flag",
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
				Key:        "dark",
				Attachment: map[string]json.RawMessage{"value": raw("dark-theme")},
			},
		},
	}

	mockStorage.AddFlag(testFlag)

	c, _ := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)

	ctx := context.Background()
	result := c.EvaluateString(ctx, "string-flag", domain.EvaluationContext{}, "light")

	assert.Equal(t, "dark-theme", result)
}

func TestCache_InvalidateFlag(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = c.InvalidateFlag(ctx, "test-flag")

	assert.NoError(t, err)
	mockStorage.AssertCalled(t, "Delete", 1)
}

func TestCache_FallbackStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		expected bool
	}{
		{"fail_open", "fail_open", true},
		{"fail_closed", "fail_closed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFlagr := flagr.NewMockClient()
			mockStorage := storage.NewMockStorage()
			eval := evaluator.New()

			c, err := New(
				WithFlagrClient(mockFlagr),
				WithStorage(mockStorage),
				WithEvaluator(eval),
				WithFallbackStrategy(tt.strategy),
			)
			require.NoError(t, err)

			ctx := context.Background()
			result := c.EvaluateBool(ctx, "nonexistent-flag", domain.EvaluationContext{})

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCache_CircuitBreaker(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	mockFlagr.GetAllFlagsFunc = func(ctx context.Context) ([]domain.Flag, error) {
		return nil, assert.AnError
	}

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
		WithCircuitBreaker(3, 1*time.Second),
	)
	require.NoError(t, err)

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		c.refreshFlags(ctx)
	}

	metrics := c.GetMetrics()
	assert.True(t, metrics.CircuitOpen)
	assert.Equal(t, 3, metrics.ConsecutiveFails)
}

func TestNewSimple(t *testing.T) {
	config := SimpleConfig{
		FlagrEndpoint:   "http://localhost:18000",
		RefreshInterval: 5 * time.Minute,
	}

	c, err := NewSimple(config)

	require.NoError(t, err)
	assert.NotNil(t, c)
}

func BenchmarkCache_Evaluate_Local(b *testing.B) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	testFlag := domain.Flag{
		ID:      1,
		Key:     "bench-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{Property: "country", Operator: domain.OperatorEQ, Value: "BR"},
				},
				Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}},
			},
		},
		Variants: []domain.Variant{
			{
				ID:         1,
				Key:        "on",
				Attachment: map[string]json.RawMessage{"enabled": raw(true)},
			},
		},
	}

	mockStorage.AddFlag(testFlag)

	c, _ := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
	)

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context:  map[string]interface{}{"country": "BR"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Evaluate(ctx, "bench-flag", evalCtx)
	}
}
