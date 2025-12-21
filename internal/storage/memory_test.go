package storage

import (
	"context"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig() Config {
	return Config{
		MaxCost:        10 * 1024 * 1024, // 10MB
		NumCounters:    10_000,           // mínimo recomendado
		BufferItems:    64,               // padrão da Ristretto
		DefaultTTL:     time.Minute,
		MetricsEnabled: false,
	}
}

func TestMemoryStorage_SetAndGet(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	flag := domain.Flag{Key: "flag1", Enabled: true}

	err = s.Set(ctx, "flag1", flag, time.Minute)
	require.NoError(t, err)

	result, err := s.Get(ctx, "flag1")
	require.NoError(t, err)
	assert.Equal(t, "flag1", result.Key)
	assert.True(t, result.Enabled)
}

func TestMemoryStorage_GetMissing(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = s.Get(ctx, "missing")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStorage_Expiration(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	flag := domain.Flag{Key: "temp"}

	err = s.Set(ctx, "temp", flag, 10*time.Millisecond)
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)

	_, err = s.Get(ctx, "temp")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStorage_Delete(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	s.Set(ctx, "a", domain.Flag{Key: "a"}, time.Minute)
	err = s.Delete(ctx, "a")
	require.NoError(t, err)

	_, err = s.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStorage_Clear(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	s.Set(ctx, "a", domain.Flag{Key: "a"}, time.Minute)
	s.Set(ctx, "b", domain.Flag{Key: "b"}, time.Minute)

	err = s.Clear(ctx)
	require.NoError(t, err)

	_, err = s.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStorage_ContextCancellation_Get(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = s.Get(ctx, "key")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMemoryStorage_ContextCancellation_Set(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = s.Set(ctx, "key", domain.Flag{Key: "test"}, time.Minute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMemoryStorage_DefaultTTL(t *testing.T) {
	cfg := newTestConfig()
	cfg.DefaultTTL = 50 * time.Millisecond
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	flag := domain.Flag{Key: "ttl-test"}

	// Set with zero TTL (should use default)
	err = s.Set(ctx, "ttl-test", flag, 0)
	require.NoError(t, err)

	// Should exist immediately
	_, err = s.Get(ctx, "ttl-test")
	require.NoError(t, err)

	// Wait for default TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be gone
	_, err = s.Get(ctx, "ttl-test")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStorage_List(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// List is not supported by ristretto
	keys, err := s.List(ctx)
	assert.Error(t, err)
	assert.Nil(t, keys)
	assert.Contains(t, err.Error(), "not supported")
}

func TestMemoryStorage_Metrics(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	metrics := s.Metrics()
	assert.Equal(t, uint64(0), metrics.KeysAdded)

	// Add a key
	s.Set(ctx, "key1", domain.Flag{Key: "key1"}, time.Minute)
	metrics = s.Metrics()
	assert.Equal(t, uint64(1), metrics.KeysAdded)
	assert.Equal(t, uint64(0), metrics.GetsKept)

	// Get the key
	s.Get(ctx, "key1")
	metrics = s.Metrics()
	assert.Equal(t, uint64(1), metrics.GetsKept)

	// Get missing key
	s.Get(ctx, "missing")
	metrics = s.Metrics()
	assert.Equal(t, uint64(1), metrics.GetsDropped)

	// Delete a key
	s.Delete(ctx, "key1")
	metrics = s.Metrics()
	assert.Equal(t, uint64(1), metrics.KeysDeleted)
}

func TestMemoryStorage_Close(t *testing.T) {
	cfg := newTestConfig()
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	err = s.Close()
	assert.NoError(t, err)
}

func TestMemoryStorage_SetRejection(t *testing.T) {
	// Create a very small cache that will reject entries
	cfg := Config{
		MaxCost:     1,     // Very small
		NumCounters: 10,    // Very small
		BufferItems: 1,     // Very small
		DefaultTTL:  time.Minute,
	}
	s, err := NewMemoryStorage(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to add multiple large items
	for i := 0; i < 100; i++ {
		flag := domain.Flag{Key: "test"}
		_ = s.Set(ctx, "key", flag, time.Minute)
	}

	// Check that some sets might have been rejected
	metrics := s.Metrics()
	// At least some keys should have been added
	assert.Greater(t, metrics.KeysAdded+metrics.SetsRejected, uint64(0))
}

func TestMockStorage_AllMethods(t *testing.T) {
	mock := NewMockStorage()
	ctx := context.Background()

	// Set behavior
	mock.SetFunc = func(ctx context.Context, key string, f domain.Flag, ttl time.Duration) error {
		if key == "fail" {
			return assert.AnError
		}
		return nil
	}

	// Get behavior
	mock.GetFunc = func(ctx context.Context, key string) (*domain.Flag, error) {
		if key == "exists" {
			return &domain.Flag{Key: "exists"}, nil
		}
		return nil, ErrNotFound
	}

	mock.DeleteFunc = func(ctx context.Context, key string) error {
		return nil
	}

	mock.ClearFunc = func(ctx context.Context) error {
		return nil
	}

	// Set success
	err := mock.Set(ctx, "ok", domain.Flag{Key: "ok"}, time.Minute)
	require.NoError(t, err)

	// Set error
	err = mock.Set(ctx, "fail", domain.Flag{Key: "x"}, time.Minute)
	require.Error(t, err)

	// Get success
	f, err := mock.Get(ctx, "exists")
	require.NoError(t, err)
	assert.Equal(t, "exists", f.Key)

	// Get not found
	_, err = mock.Get(ctx, "missing")
	assert.ErrorIs(t, err, ErrNotFound)

	// Delete
	require.NoError(t, mock.Delete(ctx, "anything"))

	// Clear
	require.NoError(t, mock.Clear(ctx))
}
