package vexilla

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFacade_NewClient tests client creation with various options.
func TestFacade_NewClient(t *testing.T) {
	tests := []struct {
		name      string
		options   []Option
		expectErr bool
	}{
		{
			name: "valid configuration",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithRefreshInterval(5 * time.Minute),
			},
			expectErr: false,
		},
		{
			name: "with all options",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithFlagrAPIKey("test-key"),
				WithFlagrTimeout(10 * time.Second),
				WithFlagrMaxRetries(5),
				WithRefreshInterval(5 * time.Minute),
				WithInitialTimeout(10 * time.Second),
				WithFallbackStrategy("fail_open"),
				WithCircuitBreaker(5, 30*time.Second),
				WithOnlyEnabled(true),
			},
			expectErr: false,
		},
		{
			name: "with service tag",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithServiceTag("test-service"),
				WithOnlyEnabled(true),
			},
			expectErr: false,
		},
		{
			name: "with additional tags",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithAdditionalTags([]string{"production", "critical"}, "any"),
			},
			expectErr: false,
		},
		{
			name: "invalid fallback strategy",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithFallbackStrategy("invalid"),
			},
			expectErr: true,
		},
		{
			name: "invalid tag match mode",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithAdditionalTags([]string{"tag1"}, "invalid"),
			},
			expectErr: true,
		},
		{
			name: "missing flagr endpoint",
			options: []Option{
				WithRefreshInterval(5 * time.Minute),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.options...)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestFacade_WithConfig tests using a Config struct.
func TestFacade_WithConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Flagr.Endpoint = "http://localhost:18000"
	cfg.Cache.RefreshInterval = 10 * time.Minute
	cfg.Cache.Filter.OnlyEnabled = true
	cfg.Cache.Filter.ServiceName = "test-service"
	cfg.Cache.Filter.RequireServiceTag = true

	client, err := New(WithConfig(cfg))

	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestFacade_Context tests the Context type and its methods.
func TestFacade_Context(t *testing.T) {
	t.Run("NewContext", func(t *testing.T) {
		ctx := NewContext("user-123")

		assert.Equal(t, "user-123", ctx.EntityID)
		assert.Equal(t, "user", ctx.EntityType)
		assert.NotNil(t, ctx.Attributes)
	})

	t.Run("WithAttribute", func(t *testing.T) {
		ctx := NewContext("user-123").
			WithAttribute("country", "BR").
			WithAttribute("tier", "premium")

		assert.Equal(t, "BR", ctx.Attributes["country"])
		assert.Equal(t, "premium", ctx.Attributes["tier"])
	})

	t.Run("WithEntityType", func(t *testing.T) {
		ctx := NewContext("device-456").
			WithEntityType("device")

		assert.Equal(t, "device", ctx.EntityType)
	})

	t.Run("fluent interface", func(t *testing.T) {
		ctx := NewContext("user-789").
			WithAttribute("country", "US").
			WithAttribute("age", 25).
			WithEntityType("user")

		assert.Equal(t, "user-789", ctx.EntityID)
		assert.Equal(t, "user", ctx.EntityType)
		assert.Equal(t, "US", ctx.Attributes["country"])
		assert.Equal(t, 25, ctx.Attributes["age"])
	})
}

// TestFacade_Result tests the Result type and its helper methods.
func TestFacade_Result(t *testing.T) {
	t.Run("IsEnabled - from variant key", func(t *testing.T) {
		tests := []struct {
			variantKey string
			expected   bool
		}{
			{"enabled", true},
			{"on", true},
			{"true", true},
			{"disabled", false},
			{"off", false},
		}

		for _, tt := range tests {
			t.Run(tt.variantKey, func(t *testing.T) {
				result := &Result{VariantKey: tt.variantKey}
				assert.Equal(t, tt.expected, result.IsEnabled())
			})
		}
	})

	t.Run("IsEnabled - from attachment", func(t *testing.T) {
		result := &Result{
			VariantAttachment: map[string]json.RawMessage{
				"enabled": json.RawMessage(`true`),
			},
		}
		assert.True(t, result.IsEnabled())

		result = &Result{
			VariantAttachment: map[string]json.RawMessage{
				"enabled": json.RawMessage(`false`),
			},
		}
		assert.False(t, result.IsEnabled())
	})

	t.Run("GetString", func(t *testing.T) {
		result := &Result{
			VariantAttachment: map[string]json.RawMessage{
				"theme": json.RawMessage(`"dark"`),
			},
		}

		assert.Equal(t, "dark", result.GetString("theme", "light"))
		assert.Equal(t, "light", result.GetString("missing", "light"))
	})

	t.Run("GetInt", func(t *testing.T) {
		result := &Result{
			VariantAttachment: map[string]json.RawMessage{
				"limit": json.RawMessage(`100`),
			},
		}

		assert.Equal(t, 100, result.GetInt("limit", 10))
		assert.Equal(t, 10, result.GetInt("missing", 10))
	})

	t.Run("GetInt - from float", func(t *testing.T) {
		result := &Result{
			VariantAttachment: map[string]json.RawMessage{
				"limit": json.RawMessage(`42.7`),
			},
		}

		assert.Equal(t, 42, result.GetInt("limit", 10))
	})
}

// TestFacade_Errors tests error types.
func TestFacade_Errors(t *testing.T) {
	t.Run("EvaluationError", func(t *testing.T) {
		err := &EvaluationError{
			FlagKey: "test-flag",
			Reason:  "test reason",
			Err:     assert.AnError,
		}

		assert.Contains(t, err.Error(), "test-flag")
		assert.Contains(t, err.Error(), "test reason")
		assert.Equal(t, assert.AnError, err.Unwrap())
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := &NotFoundError{FlagKey: "missing-flag"}

		assert.Contains(t, err.Error(), "missing-flag")
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("CircuitOpenError", func(t *testing.T) {
		err := &CircuitOpenError{Message: "too many failures"}

		assert.Contains(t, err.Error(), "circuit breaker")
		assert.Contains(t, err.Error(), "too many failures")
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := &ConfigError{
			Field:   "endpoint",
			Message: "cannot be empty",
		}

		assert.Contains(t, err.Error(), "endpoint")
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

// TestFacade_DefaultConfig tests the default configuration.
func TestFacade_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 5*time.Second, cfg.Flagr.Timeout)
	assert.Equal(t, 3, cfg.Flagr.MaxRetries)
	assert.Equal(t, 5*time.Minute, cfg.Cache.RefreshInterval)
	assert.Equal(t, 10*time.Second, cfg.Cache.InitialTimeout)
	assert.True(t, cfg.Cache.Filter.OnlyEnabled)
	assert.Equal(t, "any", cfg.Cache.Filter.TagMatchMode)
	assert.Equal(t, 3, cfg.CircuitBreaker.Threshold)
	assert.Equal(t, 30*time.Second, cfg.CircuitBreaker.Timeout)
	assert.Equal(t, "fail_closed", cfg.FallbackStrategy)
}

// TestFacade_OptionValidation tests option validation.
func TestFacade_OptionValidation(t *testing.T) {
	t.Run("empty endpoint", func(t *testing.T) {
		_, err := New(WithFlagrEndpoint(""))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint")
	})

	t.Run("invalid fallback strategy", func(t *testing.T) {
		_, err := New(
			WithFlagrEndpoint("http://localhost:18000"),
			WithFallbackStrategy("invalid_strategy"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fallback strategy")
	})

	t.Run("invalid tag match mode", func(t *testing.T) {
		_, err := New(
			WithFlagrEndpoint("http://localhost:18000"),
			WithAdditionalTags([]string{"tag1"}, "invalid_mode"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tag match mode")
	})
}

// TestFacade_MultipleOptions tests applying multiple options.
func TestFacade_MultipleOptions(t *testing.T) {
	client, err := New(
		WithFlagrEndpoint("http://localhost:18000"),
		WithFlagrAPIKey("test-key"),
		WithFlagrTimeout(15*time.Second),
		WithFlagrMaxRetries(10),
		WithRefreshInterval(10*time.Minute),
		WithInitialTimeout(20*time.Second),
		WithFallbackStrategy("fail_open"),
		WithCircuitBreaker(10, 60*time.Second),
		WithOnlyEnabled(true),
		WithServiceTag("test-service"),
		WithAdditionalTags([]string{"production"}, "any"),
	)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.cache)
}

// TestFacade_ContextConversion tests internal context conversion.
func TestFacade_ContextConversion(t *testing.T) {
	ctx := NewContext("user-123").
		WithAttribute("country", "BR").
		WithAttribute("tier", "premium").
		WithEntityType("user")

	domainCtx := toDomainContext(ctx)

	assert.Equal(t, "user-123", domainCtx.EntityID)
	assert.Equal(t, "user", domainCtx.EntityType)
	assert.Equal(t, "BR", domainCtx.Context["country"])
	assert.Equal(t, "premium", domainCtx.Context["tier"])
}

// BenchmarkFacade_NewContext benchmarks context creation.
func BenchmarkFacade_NewContext(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewContext("user-123").
			WithAttribute("country", "BR").
			WithAttribute("tier", "premium")
	}
}

// BenchmarkFacade_ResultIsEnabled benchmarks Result.IsEnabled().
func BenchmarkFacade_ResultIsEnabled(b *testing.B) {
	result := &Result{
		VariantAttachment: map[string]json.RawMessage{
			"enabled": json.RawMessage(`true`),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.IsEnabled()
	}
}

// BenchmarkFacade_ResultGetString benchmarks Result.GetString().
func BenchmarkFacade_ResultGetString(b *testing.B) {
	result := &Result{
		VariantAttachment: map[string]json.RawMessage{
			"theme": json.RawMessage(`"dark"`),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.GetString("theme", "light")
	}
}

// TestFacade_Integration tests full integration with mock Flagr.
// This test requires a running Flagr instance or mock.
func TestFacade_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a mock Flagr server
	// For now, we skip it but provide the structure
	t.Skip("Integration test requires mock Flagr server")

	// Example integration test structure:
	/*
		// Start mock Flagr server
		mockServer := httptest.NewServer(...)
		defer mockServer.Close()

		// Create client
		client, err := New(
			WithFlagrEndpoint(mockServer.URL),
			WithOnlyEnabled(true),
		)
		require.NoError(t, err)

		// Start client
		ctx := context.Background()
		err = client.Start(ctx)
		require.NoError(t, err)
		defer client.Stop()

		// Test evaluations
		evalCtx := NewContext("user-123")
		enabled := client.Bool(ctx, "test-flag", evalCtx)
		assert.True(t, enabled)

		// Test metrics
		metrics := client.Metrics()
		assert.Greater(t, metrics.Storage.KeysAdded, uint64(0))
	*/
}

// TestFacade_ConfigCombinations tests various config combinations.
func TestFacade_ConfigCombinations(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		check   func(t *testing.T, client *Client)
	}{
		{
			name: "minimal config",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
			},
			check: func(t *testing.T, client *Client) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.cache)
			},
		},
		{
			name: "with filtering",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithOnlyEnabled(true),
				WithServiceTag("test-service"),
			},
			check: func(t *testing.T, client *Client) {
				assert.NotNil(t, client)
			},
		},
		{
			name: "with circuit breaker",
			options: []Option{
				WithFlagrEndpoint("http://localhost:18000"),
				WithCircuitBreaker(5, 45*time.Second),
			},
			check: func(t *testing.T, client *Client) {
				assert.NotNil(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.options...)
			require.NoError(t, err)
			tt.check(t, client)
		})
	}
}
