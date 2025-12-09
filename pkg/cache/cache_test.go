package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	cfg := DefaultConfig()
	cache, err := NewCache(cfg)
	require.NoError(t, err)
	require.NotNil(t, cache)
	defer cache.Close()

	assert.NotNil(t, cache.store)
	assert.NotNil(t, cache.evaluator)
}

func TestCache_SetAndGetFlag(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		ID:          1,
		Key:         "test_flag",
		Description: "Test flag",
		Enabled:     true,
		Segments: []types.Segment{
			{
				ID:             1,
				RolloutPercent: 100,
				Constraints: []types.Constraint{
					{
						Property: "country",
						Operator: "EQ",
						Value:    "BR",
					},
				},
				Distributions: []types.Distribution{
					{
						VariantKey: "enabled",
						Percent:    100,
					},
				},
			},
		},
	}

	// Set flag
	err = cache.SetFlag(flag)
	require.NoError(t, err)

	// Get flag
	retrieved, found := cache.GetFlag("test_flag")
	assert.True(t, found)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "test_flag", retrieved.Key)
	assert.Equal(t, true, retrieved.Enabled)
	assert.Equal(t, 1, len(retrieved.Segments))
}

func TestCache_GetFlag_NotFound(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag, found := cache.GetFlag("nonexistent")
	assert.False(t, found)
	assert.Nil(t, flag)
}

func TestCache_SetFlag_Nil(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	err = cache.SetFlag(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil flag")
}

func TestCache_SetFlags(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flags := []*types.Flag{
		{
			ID:      1,
			Key:     "flag1",
			Enabled: true,
		},
		{
			ID:      2,
			Key:     "flag2",
			Enabled: false,
		},
		{
			ID:      3,
			Key:     "flag3",
			Enabled: true,
		},
	}

	err = cache.SetFlags(flags)
	require.NoError(t, err)

	// Verify all flags are cached
	for _, flag := range flags {
		retrieved, found := cache.GetFlag(flag.Key)
		assert.True(t, found, "Flag %s should be found", flag.Key)
		assert.Equal(t, flag.Key, retrieved.Key)
		assert.Equal(t, flag.Enabled, retrieved.Enabled)
	}
}

func TestCache_InvalidateFlag(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		ID:      1,
		Key:     "test_flag",
		Enabled: true,
	}

	// Set and verify
	err = cache.SetFlag(flag)
	require.NoError(t, err)
	_, found := cache.GetFlag("test_flag")
	assert.True(t, found)

	// Invalidate
	cache.InvalidateFlag("test_flag")

	// Verify removal
	_, found = cache.GetFlag("test_flag")
	assert.False(t, found)
}

func TestCache_Clear(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	// Add multiple flags
	flags := []*types.Flag{
		{Key: "flag1", Enabled: true},
		{Key: "flag2", Enabled: false},
		{Key: "flag3", Enabled: true},
	}
	cache.SetFlags(flags)

	// Clear cache
	cache.Clear()

	// Verify all flags are gone
	for _, flag := range flags {
		_, found := cache.GetFlag(flag.Key)
		assert.False(t, found)
	}
}

func TestCache_GetStats(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		Key:     "test_flag",
		Enabled: true,
	}

	// Initial stats
	stats := cache.GetStats()
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Set flag
	cache.SetFlag(flag)
	stats = cache.GetStats()
	assert.Equal(t, int64(1), stats.Sets)

	// Cache hit
	cache.GetFlag("test_flag")
	stats = cache.GetStats()
	assert.Equal(t, int64(1), stats.Hits)

	// Cache miss
	cache.GetFlag("nonexistent")
	stats = cache.GetStats()
	assert.Equal(t, int64(1), stats.Misses)
}

func TestCache_HitRate(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		Key:     "test_flag",
		Enabled: true,
	}
	cache.SetFlag(flag)

	// 3 hits, 1 miss = 75% hit rate
	cache.GetFlag("test_flag")
	cache.GetFlag("test_flag")
	cache.GetFlag("test_flag")
	cache.GetFlag("nonexistent")

	hitRate := cache.HitRate()
	assert.Equal(t, 0.75, hitRate)
}

func TestCache_CanEvaluateLocally(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	tests := []struct {
		name     string
		flag     *types.Flag
		expected bool
	}{
		{
			name: "static flag - 100% rollout",
			flag: &types.Flag{
				Key: "static_flag",
				Segments: []types.Segment{
					{
						RolloutPercent: 100,
						Constraints: []types.Constraint{
							{Property: "country", Operator: "EQ", Value: "BR"},
						},
						Distributions: []types.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "dynamic flag - partial rollout",
			flag: &types.Flag{
				Key: "dynamic_flag",
				Segments: []types.Segment{
					{
						RolloutPercent: 50, // Partial rollout
						Constraints: []types.Constraint{
							{Property: "country", Operator: "EQ", Value: "US"},
						},
						Distributions: []types.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "dynamic flag - multiple distributions",
			flag: &types.Flag{
				Key: "ab_test",
				Segments: []types.Segment{
					{
						RolloutPercent: 100,
						Distributions: []types.Distribution{
							{VariantKey: "control", Percent: 50},
							{VariantKey: "variant", Percent: 50},
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.SetFlag(tt.flag)
			result := cache.CanEvaluateLocally(tt.flag.Key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCache_EvaluateLocal(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		Key: "brazil_launch",
		Segments: []types.Segment{
			{
				RolloutPercent: 100,
				Constraints: []types.Constraint{
					{Property: "country", Operator: "EQ", Value: "BR"},
				},
				Distributions: []types.Distribution{
					{VariantKey: "enabled", Percent: 100},
				},
			},
		},
	}
	cache.SetFlag(flag)

	tests := []struct {
		name     string
		ctx      types.EvaluationContext
		expected bool
	}{
		{
			name: "matches constraint",
			ctx: types.EvaluationContext{
				EntityID: "user1",
				Context: map[string]interface{}{
					"country": "BR",
				},
			},
			expected: true,
		},
		{
			name: "does not match constraint",
			ctx: types.EvaluationContext{
				EntityID: "user2",
				Context: map[string]interface{}{
					"country": "US",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cache.EvaluateLocal("brazil_launch", tt.ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCache_EvaluateLocal_NotFound(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	ctx := types.EvaluationContext{
		EntityID: "user1",
		Context:  map[string]interface{}{},
	}

	_, err = cache.EvaluateLocal("nonexistent", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.PersistenceEnabled = true
	cfg.PersistencePath = tmpDir

	cache, err := NewCache(cfg)
	require.NoError(t, err)

	// Add flags
	flag := &types.Flag{
		ID:      1,
		Key:     "persistent_flag",
		Enabled: true,
	}
	err = cache.SetFlag(flag)
	require.NoError(t, err)

	// Verify file exists
	flagPath := filepath.Join(tmpDir, "persistent_flag.json")
	_, err = os.Stat(flagPath)
	assert.NoError(t, err)

	// Close and reopen cache
	cache.Close()

	cache2, err := NewCache(cfg)
	require.NoError(t, err)
	defer cache2.Close()

	// Wait for Ristretto to process
	time.Sleep(100 * time.Millisecond)

	// Verify flag was restored
	restored, found := cache2.GetFlag("persistent_flag")
	assert.True(t, found)
	if found {
		assert.Equal(t, "persistent_flag", restored.Key)
		assert.Equal(t, true, restored.Enabled)
	}
}

func TestCache_WarmUp(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flags := []*types.Flag{
		{Key: "flag1", Enabled: true},
		{Key: "flag2", Enabled: false},
		{Key: "flag3", Enabled: true},
	}

	ctx := context.Background()
	err = cache.WarmUp(ctx, flags)
	require.NoError(t, err)

	// Verify all flags are cached
	for _, flag := range flags {
		retrieved, found := cache.GetFlag(flag.Key)
		assert.True(t, found)
		assert.Equal(t, flag.Key, retrieved.Key)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache, err := NewCache(DefaultConfig())
	require.NoError(t, err)
	defer cache.Close()

	flag := &types.Flag{
		Key:     "concurrent_flag",
		Enabled: true,
	}
	cache.SetFlag(flag)

	// Concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				// Read
				cache.GetFlag("concurrent_flag")

				// Write
				cache.SetFlag(&types.Flag{
					Key:     "concurrent_flag",
					Enabled: j%2 == 0,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should still be functional
	_, found := cache.GetFlag("concurrent_flag")
	assert.True(t, found)
}

func BenchmarkCache_GetFlag(b *testing.B) {
	cache, _ := NewCache(DefaultConfig())
	defer cache.Close()

	flag := &types.Flag{
		Key:     "benchmark_flag",
		Enabled: true,
	}
	cache.SetFlag(flag)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetFlag("benchmark_flag")
	}
}

func BenchmarkCache_SetFlag(b *testing.B) {
	cache, _ := NewCache(DefaultConfig())
	defer cache.Close()

	flag := &types.Flag{
		Key:     "benchmark_flag",
		Enabled: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.SetFlag(flag)
	}
}

func BenchmarkCache_EvaluateLocal(b *testing.B) {
	cache, _ := NewCache(DefaultConfig())
	defer cache.Close()

	flag := &types.Flag{
		Key: "benchmark_flag",
		Segments: []types.Segment{
			{
				RolloutPercent: 100,
				Constraints: []types.Constraint{
					{Property: "country", Operator: "EQ", Value: "BR"},
				},
				Distributions: []types.Distribution{
					{VariantKey: "enabled", Percent: 100},
				},
			},
		},
	}
	cache.SetFlag(flag)

	ctx := types.EvaluationContext{
		EntityID: "user1",
		Context: map[string]interface{}{
			"country": "BR",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.EvaluateLocal("benchmark_flag", ctx)
	}
}
