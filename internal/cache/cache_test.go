package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
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

func TestCache_CircuitBreaker_CustomThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		failures  int
		wantOpen  bool
	}{
		{
			name:      "circuit opens at custom threshold of 5",
			threshold: 5,
			failures:  5,
			wantOpen:  true,
		},
		{
			name:      "circuit stays closed below threshold",
			threshold: 5,
			failures:  4,
			wantOpen:  false,
		},
		{
			name:      "circuit opens at threshold of 2",
			threshold: 2,
			failures:  2,
			wantOpen:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				WithCircuitBreaker(tt.threshold, 1*time.Second),
			)
			require.NoError(t, err)

			ctx := context.Background()

			// Trigger failures
			for i := 0; i < tt.failures; i++ {
				c.refreshFlags(ctx)
			}

			metrics := c.GetMetrics()
			assert.Equal(t, tt.wantOpen, metrics.CircuitOpen, "circuit breaker state mismatch")
			assert.Equal(t, tt.failures, metrics.ConsecutiveFails)
		})
	}
}

func TestCache_CircuitBreaker_AutoReset(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	failCount := 0
	mockFlagr.GetAllFlagsFunc = func(ctx context.Context) ([]domain.Flag, error) {
		failCount++
		return nil, assert.AnError
	}

	// Use a very short timeout for testing
	timeout := 150 * time.Millisecond
	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
		WithCircuitBreaker(2, timeout),
	)
	require.NoError(t, err)

	// Trigger failures to open circuit using handleRefreshError
	c.handleRefreshError(assert.AnError)
	c.handleRefreshError(assert.AnError)

	metrics := c.GetMetrics()
	assert.True(t, metrics.CircuitOpen, "circuit should be open after threshold failures")
	assert.Equal(t, 2, metrics.ConsecutiveFails)

	// Wait for auto-reset with some margin
	time.Sleep(timeout + 100*time.Millisecond)

	// Verify circuit has been reset (only circuitOpen flag is reset, not consecutiveFails)
	c.mu.RLock()
	circuitState := c.circuitOpen
	c.mu.RUnlock()

	assert.False(t, circuitState, "circuit should be closed after timeout")
}

func TestCache_CircuitBreaker_HandleRefreshError(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	threshold := 4
	timeout := 50 * time.Millisecond

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
		WithCircuitBreaker(threshold, timeout),
	)
	require.NoError(t, err)

	// Simulate consecutive failures using handleRefreshError
	for i := 0; i < threshold-1; i++ {
		c.handleRefreshError(assert.AnError)
	}

	metrics := c.GetMetrics()
	assert.False(t, metrics.CircuitOpen, "circuit should still be closed")
	assert.Equal(t, threshold-1, metrics.ConsecutiveFails)

	// One more failure should open the circuit
	c.handleRefreshError(assert.AnError)

	metrics = c.GetMetrics()
	assert.True(t, metrics.CircuitOpen, "circuit should be open")
	assert.Equal(t, threshold, metrics.ConsecutiveFails)

	// Wait for timeout and verify auto-reset
	time.Sleep(timeout + 20*time.Millisecond)

	metrics = c.GetMetrics()
	assert.False(t, metrics.CircuitOpen, "circuit should auto-reset after timeout")
}

func TestCache_CircuitBreaker_CustomTimeoutRespected(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		timeout   time.Duration
	}{
		{
			name:      "custom timeout 100ms",
			threshold: 2,
			timeout:   100 * time.Millisecond,
		},
		{
			name:      "custom timeout 200ms",
			threshold: 3,
			timeout:   200 * time.Millisecond,
		},
		{
			name:      "custom timeout 50ms",
			threshold: 2,
			timeout:   50 * time.Millisecond,
		},
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
				WithCircuitBreaker(tt.threshold, tt.timeout),
			)
			require.NoError(t, err)

			// Verify config values are set correctly
			assert.Equal(t, tt.threshold, c.config.CircuitBreakerThreshold)
			assert.Equal(t, tt.timeout, c.config.CircuitBreakerTimeout)

			// Trigger failures to open circuit
			for i := 0; i < tt.threshold; i++ {
				c.handleRefreshError(assert.AnError)
			}

			// Verify circuit is open
			metrics := c.GetMetrics()
			assert.True(t, metrics.CircuitOpen, "circuit should be open after threshold failures")

			// Wait for timeout with small margin
			time.Sleep(tt.timeout + 50*time.Millisecond)

			// Verify circuit has been reset after the configured timeout
			c.mu.RLock()
			circuitState := c.circuitOpen
			c.mu.RUnlock()

			assert.False(t, circuitState, "circuit should be closed after configured timeout")
		})
	}
}

func TestCache_CircuitBreaker_BackwardCompatibility(t *testing.T) {
	t.Run("zero threshold falls back to 3", func(t *testing.T) {
		mockFlagr := flagr.NewMockClient()
		mockStorage := storage.NewMockStorage()
		eval := evaluator.New()

		// Create cache with zero threshold (should fallback to 3)
		c := &Cache{
			flagrClient: mockFlagr,
			storage:     mockStorage,
			evaluator:   eval,
			config: Config{
				RefreshInterval:         5 * time.Minute,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 0, // This should trigger fallback
				CircuitBreakerTimeout:   100 * time.Millisecond,
			},
		}

		// Trigger 2 failures - should not open (threshold is 3)
		c.handleRefreshError(assert.AnError)
		c.handleRefreshError(assert.AnError)

		metrics := c.GetMetrics()
		assert.False(t, metrics.CircuitOpen, "circuit should not open with 2 failures when threshold is 3")

		// Trigger 3rd failure - should open
		c.handleRefreshError(assert.AnError)

		metrics = c.GetMetrics()
		assert.True(t, metrics.CircuitOpen, "circuit should open with 3 failures")
	})

	t.Run("zero timeout falls back to 30s", func(t *testing.T) {
		mockFlagr := flagr.NewMockClient()
		mockStorage := storage.NewMockStorage()
		eval := evaluator.New()

		// Create cache with zero timeout (should fallback to 30s)
		c := &Cache{
			flagrClient: mockFlagr,
			storage:     mockStorage,
			evaluator:   eval,
			config: Config{
				RefreshInterval:         5 * time.Minute,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 2,
				CircuitBreakerTimeout:   0, // This should trigger fallback to 30s
			},
		}

		// Trigger failures to open circuit
		c.handleRefreshError(assert.AnError)
		c.handleRefreshError(assert.AnError)

		metrics := c.GetMetrics()
		assert.True(t, metrics.CircuitOpen, "circuit should be open")

		// Wait for a short time (much less than 30s)
		time.Sleep(100 * time.Millisecond)

		// Circuit should still be open (30s timeout hasn't elapsed)
		c.mu.RLock()
		circuitState := c.circuitOpen
		c.mu.RUnlock()

		assert.True(t, circuitState, "circuit should still be open after 100ms with 30s timeout")
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name:      "valid default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid refresh interval",
			config: Config{
				RefreshInterval:         0,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 3,
				CircuitBreakerTimeout:   30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker threshold",
			config: Config{
				RefreshInterval:         5 * time.Minute,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 0,
				CircuitBreakerTimeout:   30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid circuit breaker timeout",
			config: Config{
				RefreshInterval:         5 * time.Minute,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 3,
				CircuitBreakerTimeout:   0,
			},
			wantError: true,
		},
		{
			name: "negative circuit breaker timeout",
			config: Config{
				RefreshInterval:         5 * time.Minute,
				InitialTimeout:          10 * time.Second,
				FallbackStrategy:        "fail_closed",
				CircuitBreakerThreshold: 3,
				CircuitBreakerTimeout:   -1 * time.Second,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
