package storage

import (
	"context"
	"sync"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/dgraph-io/ristretto"
)

// MemoryStore wraps Ristretto for flag storage
type MemoryStore struct {
	cache *ristretto.Cache
	mu    sync.RWMutex
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(maxCost int64, numCounters int64, bufferItems int64) (*MemoryStore, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	})
	if err != nil {
		return nil, err
	}

	return &MemoryStore{
		cache: cache,
	}, nil
}

// Get retrieves a flag by key
func (m *MemoryStore) Get(ctx context.Context, key string) (*vexilla.Flag, bool) {
	value, found := m.cache.Get(key)
	if !found {
		return nil, false
	}

	if flag, ok := value.(vexilla.Flag); ok {
		return &flag, true
	}

	return nil, false
}

// Set stores a flag
func (m *MemoryStore) Set(ctx context.Context, key string, flag vexilla.Flag) bool {
	return m.cache.Set(key, flag, 1)
}

// Delete removes a flag
func (m *MemoryStore) Delete(ctx context.Context, key string) {
	m.cache.Del(key)
}

// Clear removes all flags
func (m *MemoryStore) Clear(ctx context.Context) {
	m.cache.Clear()
}

// Wait waits for pending writes to complete
func (m *MemoryStore) Wait() {
	m.cache.Wait()
}

// Close closes the store
func (m *MemoryStore) Close() {
	m.cache.Close()
}

// Metrics returns cache metrics
func (m *MemoryStore) Metrics() *ristretto.Metrics {
	return m.cache.Metrics
}
