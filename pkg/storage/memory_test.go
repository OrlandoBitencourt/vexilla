package storage

import (
	"context"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// ---------------------------------------------------------
// MemoryStorage Tests
// ---------------------------------------------------------
//

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

//
// ---------------------------------------------------------
// MockStorage Tests
// ---------------------------------------------------------
//

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
