package vexilla

import (
	"context"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/internal/cache"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheAdapter_GetMetrics(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := cache.New(
		cache.WithFlagrClient(mockFlagr),
		cache.WithStorage(mockStorage),
		cache.WithEvaluator(eval),
	)
	require.NoError(t, err)

	adapter := &cacheAdapter{cache: c}

	// GetMetrics should return interface{}
	metrics := adapter.GetMetrics()
	assert.NotNil(t, metrics)

	// Should be convertible to cache.Metrics
	cacheMetrics, ok := metrics.(cache.Metrics)
	assert.True(t, ok, "metrics should be convertible to cache.Metrics")
	assert.NotNil(t, cacheMetrics)
}

func TestCacheAdapter_InvalidateFlag(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := cache.New(
		cache.WithFlagrClient(mockFlagr),
		cache.WithStorage(mockStorage),
		cache.WithEvaluator(eval),
	)
	require.NoError(t, err)

	adapter := &cacheAdapter{cache: c}

	// Should not panic
	err = adapter.InvalidateFlag("test-flag")
	assert.NoError(t, err)

	// Verify storage was called
	mockStorage.AssertCalled(t, "Delete", 1)
}

func TestCacheAdapter_InvalidateAll(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := cache.New(
		cache.WithFlagrClient(mockFlagr),
		cache.WithStorage(mockStorage),
		cache.WithEvaluator(eval),
	)
	require.NoError(t, err)

	adapter := &cacheAdapter{cache: c}

	// Should not panic
	err = adapter.InvalidateAll()
	assert.NoError(t, err)

	// Verify storage was called
	mockStorage.AssertCalled(t, "Clear", 1)
}

func TestCacheAdapter_RefreshFlags(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := cache.New(
		cache.WithFlagrClient(mockFlagr),
		cache.WithStorage(mockStorage),
		cache.WithEvaluator(eval),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)
	defer c.Stop()

	adapter := &cacheAdapter{cache: c}

	// Should not panic
	err = adapter.RefreshFlags()
	// May return error if no flags, but should not panic
	_ = err
}

func TestCacheAdapter_ImplementsInterface(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	c, err := cache.New(
		cache.WithFlagrClient(mockFlagr),
		cache.WithStorage(mockStorage),
		cache.WithEvaluator(eval),
	)
	require.NoError(t, err)

	adapter := &cacheAdapter{cache: c}

	// Verify it has all required methods
	assert.NotNil(t, adapter.GetMetrics)
	assert.NotNil(t, adapter.InvalidateFlag)
	assert.NotNil(t, adapter.InvalidateAll)
	assert.NotNil(t, adapter.RefreshFlags)
}
