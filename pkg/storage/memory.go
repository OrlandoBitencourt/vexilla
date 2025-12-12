package storage

import (
	"context"
	"errors"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
	"github.com/dgraph-io/ristretto"
)

type MemoryStorage struct {
	cache   *ristretto.Cache
	config  Config
	metrics Metrics
}

func NewMemoryStorage(cfg Config) (*MemoryStorage, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: int64(cfg.BufferItems),
	})
	if err != nil {
		return nil, err
	}

	return &MemoryStorage{
		cache:  cache,
		config: cfg,
	}, nil
}

func (m *MemoryStorage) Get(ctx context.Context, key string) (*domain.Flag, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	val, ok := m.cache.Get(key)
	if !ok {
		m.metrics.GetsDropped++
		return nil, ErrNotFound
	}

	flag := val.(domain.Flag)
	m.metrics.GetsKept++
	return &flag, nil
}

func (m *MemoryStorage) Set(ctx context.Context, key string, flag domain.Flag, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if ttl == 0 {
		ttl = m.config.DefaultTTL
	}

	ok := m.cache.SetWithTTL(key, flag, 1, ttl)
	if !ok {
		m.metrics.SetsRejected++
		return errors.New("cache rejected set")
	}

	m.metrics.KeysAdded++
	m.cache.Wait()

	return nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	m.cache.Del(key)
	m.metrics.KeysDeleted++
	return nil
}

func (m *MemoryStorage) Clear(ctx context.Context) error {
	m.cache.Clear()
	return nil
}

func (m *MemoryStorage) List(ctx context.Context) ([]string, error) {
	// Ristretto não fornece listagem; manteríamos mapa manual — opcional.
	return nil, errors.New("memory storage cannot list keys (not supported by ristretto)")
}

func (m *MemoryStorage) Metrics() Metrics { return m.metrics }

func (m *MemoryStorage) Close() error { m.cache.Close(); return nil }
