package vexilla

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig tests the default configuration
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 5*time.Minute, config.RefreshInterval)
	assert.Equal(t, 10*time.Second, config.InitialTimeout)
	assert.Equal(t, 5*time.Second, config.HTTPTimeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, "fail_closed", config.FallbackStrategy)
	assert.True(t, config.PersistenceEnabled)
	assert.False(t, config.WebhookEnabled)
	assert.False(t, config.AdminAPIEnabled)
	assert.Equal(t, int64(1<<30), config.CacheMaxCost)
	assert.Equal(t, int64(1e7), config.CacheNumCounters)
}

// TestConfigValidate tests configuration validation
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty endpoint",
			config: Config{
				FlagrEndpoint:   "",
				RefreshInterval: 1 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "negative refresh interval",
			config: Config{
				FlagrEndpoint:   "http://localhost",
				RefreshInterval: -1 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid fallback strategy",
			config: Config{
				FlagrEndpoint:    "http://localhost",
				RefreshInterval:  1 * time.Minute,
				FallbackStrategy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNew creates a new cache instance
func TestNew(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, cache)

	assert.NotNil(t, cache.cache)
	assert.NotNil(t, cache.httpClient)
	assert.NotNil(t, cache.evaluator)
	assert.NotNil(t, cache.tracer)
	assert.NotNil(t, cache.meter)

	// Cleanup
	cache.Stop()
}

// TestNewWithInvalidConfig tests cache creation with invalid config
func TestNewWithInvalidConfig(t *testing.T) {
	config := Config{
		FlagrEndpoint: "", // Invalid
	}

	cache, err := New(config)
	assert.Error(t, err)
	assert.Nil(t, cache)
}

// TestCacheStartStop tests starting and stopping the cache
func TestCacheStartStop(t *testing.T) {
	// Mock Flagr server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Flag{})
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false
	config.WebhookEnabled = false
	config.AdminAPIEnabled = false

	cache, err := New(config)
	require.NoError(t, err)

	// Start should succeed
	err = cache.Start()
	assert.NoError(t, err)

	// Give it a moment to initialize
	time.Sleep(100 * time.Millisecond)

	// Stop should succeed
	err = cache.Stop()
	assert.NoError(t, err)
}

// TestEvaluateBool tests boolean evaluation
func TestEvaluateBool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/flags" {
			flags := []Flag{
				{
					ID:      1,
					Key:     "test_flag",
					Default: true,
					Rules:   []FlagRule{},
				},
			}
			json.NewEncoder(w).Encode(flags)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	require.NoError(t, cache.Start())
	defer cache.Stop()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	evalCtx := EvaluationContext{
		UserID:     "user123",
		Attributes: map[string]interface{}{"country": "BR"},
	}

	result := cache.EvaluateBool(ctx, "test_flag", evalCtx)
	assert.True(t, result)
}

// TestEvaluateString tests string evaluation
func TestEvaluateString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/flags" {
			flags := []Flag{
				{
					Key:     "theme",
					Default: "dark",
				},
			}
			json.NewEncoder(w).Encode(flags)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	require.NoError(t, cache.Start())
	defer cache.Stop()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	result := cache.EvaluateString(ctx, "theme", EvaluationContext{}, "light")
	assert.Equal(t, "dark", result)
}

// TestEvaluateInt tests integer evaluation
func TestEvaluateInt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/flags" {
			flags := []Flag{
				{
					Key:     "max_items",
					Default: 100,
				},
			}
			json.NewEncoder(w).Encode(flags)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	require.NoError(t, cache.Start())
	defer cache.Stop()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	result := cache.EvaluateInt(ctx, "max_items", EvaluationContext{}, 50)
	assert.Equal(t, 100, result)
}

// TestCanEvaluateLocally tests local evaluation detection
func TestCanEvaluateLocally(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	tests := []struct {
		name     string
		flag     Flag
		expected bool
	}{
		{
			name: "100% rollout - can evaluate locally",
			flag: Flag{
				Key: "static_flag",
				Segments: []Segment{
					{
						RolloutPercent: 100,
						Constraints: []Constraint{
							{Property: "country", Operator: "EQ", Value: "BR"},
						},
						Distributions: []VariantDistribution{
							{VariantKey: "enabled", Percentage: 100},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "partial rollout - needs Flagr",
			flag: Flag{
				Key: "dynamic_flag",
				Segments: []Segment{
					{
						RolloutPercent: 50, // Partial!
						Distributions: []VariantDistribution{
							{VariantKey: "enabled", Percentage: 100},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "A/B test - needs Flagr",
			flag: Flag{
				Key: "ab_test",
				Segments: []Segment{
					{
						RolloutPercent: 100,
						Distributions: []VariantDistribution{
							{VariantKey: "control", Percentage: 50},
							{VariantKey: "variant", Percentage: 50},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "percentage distribution - needs Flagr",
			flag: Flag{
				Key: "gradual",
				Segments: []Segment{
					{
						RolloutPercent: 100,
						Distributions: []VariantDistribution{
							{VariantKey: "enabled", Percentage: 75},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.canEvaluateLocally(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetermineEvaluationStrategy tests strategy determination
func TestDetermineEvaluationStrategy(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	tests := []struct {
		name     string
		flag     Flag
		expected string
	}{
		{
			name: "explicit static",
			flag: Flag{
				EvaluationType: "static",
			},
			expected: "static",
		},
		{
			name: "explicit dynamic",
			flag: Flag{
				EvaluationType: "dynamic",
			},
			expected: "dynamic",
		},
		{
			name: "auto-detect static",
			flag: Flag{
				EvaluationType: "auto",
				Segments: []Segment{
					{
						RolloutPercent: 100,
						Distributions: []VariantDistribution{
							{Percentage: 100},
						},
					},
				},
			},
			expected: "static",
		},
		{
			name: "auto-detect dynamic",
			flag: Flag{
				Segments: []Segment{
					{
						RolloutPercent: 50,
					},
				},
			},
			expected: "dynamic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.determineEvaluationStrategy(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInvalidateFlag tests flag invalidation
func TestInvalidateFlag(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	// Add a flag to cache
	flag := Flag{Key: "test", Default: true}
	cache.cache.Set("test", flag, 1)
	cache.cache.Wait()

	// Verify it exists
	_, found := cache.cache.Get("test")
	assert.True(t, found)

	// Invalidate
	ctx := context.Background()
	cache.InvalidateFlag(ctx, "test")

	// Verify it's gone
	_, found = cache.cache.Get("test")
	assert.False(t, found)
}

// TestInvalidateAll tests clearing entire cache
func TestInvalidateAll(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	// Add multiple flags
	for i := 0; i < 5; i++ {
		flag := Flag{Key: string(rune('a' + i)), Default: true}
		cache.cache.Set(flag.Key, flag, 1)
	}
	cache.cache.Wait()

	// Invalidate all
	ctx := context.Background()
	cache.InvalidateAll(ctx)

	// Verify all are gone
	for i := 0; i < 5; i++ {
		_, found := cache.cache.Get(string(rune('a' + i)))
		assert.False(t, found)
	}
}

// TestGetCacheStats tests cache statistics
func TestGetCacheStats(t *testing.T) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	stats := cache.GetCacheStats()
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.KeysAdded, uint64(0))
}

// TestPersistence tests disk persistence
func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = true
	config.PersistencePath = tmpDir

	cache, err := New(config)
	require.NoError(t, err)

	// Save to disk
	ctx := context.Background()
	err = cache.saveToDisk(ctx)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tmpDir + "/flags.json")
	assert.NoError(t, err)

	cache.Stop()
}

// TestCircuitBreaker tests circuit breaker behavior
func TestCircuitBreaker(t *testing.T) {
	// Server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false
	config.RetryAttempts = 1

	cache, err := New(config)
	require.NoError(t, err)
	defer cache.Stop()

	// Trigger multiple failures
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		go cache.refreshFlags(ctx)
	}

	// Circuit should be open
	stats := cache.GetCacheStats()
	assert.True(t, stats.CircuitOpen || stats.ConsecutiveFails > 0)
}

// TestEvaluatorBasic tests basic evaluator functionality
func TestEvaluatorBasic(t *testing.T) {
	evaluator := NewEvaluator()
	assert.NotNil(t, evaluator)

	flag := Flag{
		Key:     "test",
		Default: false,
		Rules: []FlagRule{
			{
				Condition: `user_id == "123"`,
				Value:     true,
			},
		},
	}

	ctx := context.Background()
	evalCtx := EvaluationContext{
		UserID:     "123",
		Attributes: map[string]interface{}{},
	}

	result := evaluator.Evaluate(ctx, flag, evalCtx)
	assert.Equal(t, true, result)
}

// TestEvaluatorWithAttributes tests evaluator with attributes
func TestEvaluatorWithAttributes(t *testing.T) {
	evaluator := NewEvaluator()

	flag := Flag{
		Key:     "country_check",
		Default: false,
		Rules: []FlagRule{
			{
				Condition: `country == "BR"`,
				Value:     true,
			},
		},
	}

	ctx := context.Background()
	evalCtx := EvaluationContext{
		UserID: "user1",
		Attributes: map[string]interface{}{
			"country": "BR",
		},
	}

	result := evaluator.Evaluate(ctx, flag, evalCtx)
	assert.Equal(t, true, result)
}

// BenchmarkEvaluateBool benchmarks boolean evaluation
func BenchmarkEvaluateBool(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Flag{
			{Key: "bench", Default: true},
		})
	}))
	defer server.Close()

	config := DefaultConfig()
	config.FlagrEndpoint = server.URL
	config.PersistenceEnabled = false

	cache, _ := New(config)
	cache.Start()
	defer cache.Stop()

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	evalCtx := EvaluationContext{UserID: "user123"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.EvaluateBool(ctx, "bench", evalCtx)
	}
}

// BenchmarkCanEvaluateLocally benchmarks strategy detection
func BenchmarkCanEvaluateLocally(b *testing.B) {
	config := DefaultConfig()
	config.FlagrEndpoint = "http://localhost"
	config.PersistenceEnabled = false

	cache, _ := New(config)
	defer cache.Stop()

	flag := Flag{
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []VariantDistribution{
					{Percentage: 100},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.canEvaluateLocally(flag)
	}
}
