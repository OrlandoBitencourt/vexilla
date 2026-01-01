package storage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/require"
)

// TestMemoryStorage_ConcurrentAccess tests concurrent access to metrics
// This test specifically validates that the race condition fix is working
func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)
	defer s.Close()

	ctx := context.Background()

	// Number of concurrent goroutines
	const numGoroutines = 100
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines that perform Get/Set operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				// Perform a Set operation
				flag := domain.Flag{
					Key:     "test-flag",
					Enabled: true,
				}
				_ = s.Set(ctx, "test-flag", flag, time.Minute)

				// Perform a Get operation
				_, _ = s.Get(ctx, "test-flag")

				// Try to get a non-existent key (increments GetsDropped)
				_, _ = s.Get(ctx, "missing-key")

				// Read metrics concurrently
				_ = s.Metrics()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify that metrics were updated (at least some operations succeeded)
	metrics := s.Metrics()
	t.Logf("Final metrics: KeysAdded=%d, GetsKept=%d, GetsDropped=%d",
		metrics.KeysAdded, metrics.GetsKept, metrics.GetsDropped)

	// At least some operations should have succeeded
	require.Greater(t, metrics.KeysAdded, uint64(0), "should have added keys")
	require.Greater(t, metrics.GetsKept, uint64(0), "should have successful gets")
	require.Greater(t, metrics.GetsDropped, uint64(0), "should have dropped gets")
}

// TestDiskStorage_ConcurrentAccess tests concurrent access to disk storage metrics
func TestDiskStorage_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	s, err := NewDiskStorage(dir)
	require.NoError(t, err)
	defer s.Close()

	ctx := context.Background()

	// Number of concurrent goroutines
	const numGoroutines = 50
	const opsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch concurrent goroutines that perform operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				// Perform a Set operation
				flag := domain.Flag{
					Key:     "test-flag",
					Enabled: true,
				}
				_ = s.Set(ctx, "test-flag", flag, time.Minute)

				// Perform a Get operation
				_, _ = s.Get(ctx, "test-flag")

				// Read metrics concurrently
				_ = s.Metrics()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify that metrics were updated
	metrics := s.Metrics()
	t.Logf("Final metrics: KeysAdded=%d, KeysUpdated=%d, GetsKept=%d",
		metrics.KeysAdded, metrics.KeysUpdated, metrics.GetsKept)

	// At least some operations should have succeeded
	require.Greater(t, metrics.KeysAdded+metrics.KeysUpdated, uint64(0), "should have set keys")
	require.Greater(t, metrics.GetsKept, uint64(0), "should have successful gets")
}

// TestMemoryStorage_MetricsSnapshot tests that snapshot returns consistent values
func TestMemoryStorage_MetricsSnapshot(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)
	defer s.Close()

	ctx := context.Background()

	// Perform some operations
	flag := domain.Flag{Key: "test", Enabled: true}
	_ = s.Set(ctx, "test", flag, time.Minute)
	_, _ = s.Get(ctx, "test")
	_, _ = s.Get(ctx, "missing")

	// Get metrics multiple times - should be consistent
	m1 := s.Metrics()
	m2 := s.Metrics()

	require.Equal(t, m1.KeysAdded, m2.KeysAdded)
	require.Equal(t, m1.GetsKept, m2.GetsKept)
	require.Equal(t, m1.GetsDropped, m2.GetsDropped)
}

// BenchmarkMemoryStorage_ConcurrentGets benchmarks concurrent Get operations
func BenchmarkMemoryStorage_ConcurrentGets(b *testing.B) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(b, err)
	defer s.Close()

	ctx := context.Background()

	// Pre-populate with a flag
	flag := domain.Flag{Key: "bench", Enabled: true}
	_ = s.Set(ctx, "bench", flag, time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = s.Get(ctx, "bench")
		}
	})
}

// BenchmarkMemoryStorage_ConcurrentSets benchmarks concurrent Set operations
func BenchmarkMemoryStorage_ConcurrentSets(b *testing.B) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(b, err)
	defer s.Close()

	ctx := context.Background()
	flag := domain.Flag{Key: "bench", Enabled: true}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.Set(ctx, "bench", flag, time.Hour)
		}
	})
}

// BenchmarkMemoryStorage_MetricsRead benchmarks reading metrics
func BenchmarkMemoryStorage_MetricsRead(b *testing.B) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(b, err)
	defer s.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.Metrics()
		}
	})
}
