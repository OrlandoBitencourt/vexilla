package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/types"
	"github.com/dgraph-io/ristretto"
)

// Cache implements types.FlagCache interface with Ristretto backing
type Cache struct {
	store              *ristretto.Cache
	evaluator          types.FlagEvaluator
	mu                 sync.RWMutex
	stats              Stats
	persistencePath    string
	persistenceEnabled bool
}

// Stats tracks cache performance metrics
type Stats struct {
	Hits       int64
	Misses     int64
	Sets       int64
	Evictions  int64
	LastUpdate time.Time
}

// Config holds cache configuration
type Config struct {
	MaxCost            int64
	NumCounters        int64
	PersistenceEnabled bool
	PersistencePath    string
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxCost:            1 << 30, // 1GB
		NumCounters:        1e7,     // 10M counters
		PersistenceEnabled: false,
		PersistencePath:    "/tmp/vexilla",
	}
}

// NewCache creates a new cache instance
func NewCache(cfg Config) (*Cache, error) {
	store, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: 64,
		Metrics:     true,
		OnEvict: func(item *ristretto.Item) {
			// Track evictions
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	c := &Cache{
		store:              store,
		evaluator:          evaluator.NewEvaluator(),
		persistenceEnabled: cfg.PersistenceEnabled,
		persistencePath:    cfg.PersistencePath,
		stats: Stats{
			LastUpdate: time.Now(),
		},
	}

	// Load persisted flags if enabled
	if cfg.PersistenceEnabled {
		if err := c.loadFromDisk(); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: failed to load persisted flags: %v\n", err)
		}
	}

	return c, nil
}

// GetFlag retrieves a flag from cache
func (c *Cache) GetFlag(flagKey string) (*types.Flag, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, found := c.store.Get(flagKey)
	if !found {
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	flag, ok := val.(*types.Flag)
	return flag, ok
}

// SetFlag stores a flag in cache
func (c *Cache) SetFlag(flag *types.Flag) error {
	if flag == nil {
		return fmt.Errorf("cannot cache nil flag")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate cost (approximate size in bytes)
	cost := c.calculateCost(flag)

	// Store in Ristretto
	if !c.store.Set(flag.Key, flag, cost) {
		return fmt.Errorf("failed to set flag in cache")
	}

	c.stats.Sets++
	c.stats.LastUpdate = time.Now()

	// Persist if enabled
	if c.persistenceEnabled {
		if err := c.persistFlag(flag); err != nil {
			// Log warning but don't fail the cache operation
			fmt.Printf("Warning: failed to persist flag %s: %v\n", flag.Key, err)
		}
	}

	return nil
}

// SetFlags stores multiple flags atomically
func (c *Cache) SetFlags(flags []*types.Flag) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, flag := range flags {
		if flag == nil {
			continue
		}

		cost := c.calculateCost(flag)
		c.store.Set(flag.Key, flag, cost)
		c.stats.Sets++
	}

	c.stats.LastUpdate = time.Now()

	// Persist all if enabled
	if c.persistenceEnabled {
		if err := c.persistAllFlags(flags); err != nil {
			fmt.Printf("Warning: failed to persist flags: %v\n", err)
		}
	}

	return nil
}

// InvalidateFlag removes a flag from cache
func (c *Cache) InvalidateFlag(flagKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store.Del(flagKey)

	// Remove from disk if persistence enabled
	if c.persistenceEnabled {
		path := filepath.Join(c.persistencePath, fmt.Sprintf("%s.json", flagKey))
		os.Remove(path) // Ignore errors
	}
}

// Clear removes all flags from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store.Clear()

	// Clear disk cache if enabled
	if c.persistenceEnabled {
		os.RemoveAll(c.persistencePath)
		os.MkdirAll(c.persistencePath, 0755)
	}
}

// GetStats returns current cache statistics
func (c *Cache) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := c.store.Metrics
	return Stats{
		Hits:       c.stats.Hits,
		Misses:     c.stats.Misses,
		Sets:       c.stats.Sets,
		Evictions:  int64(metrics.KeysEvicted()),
		LastUpdate: c.stats.LastUpdate,
	}
}

// HitRate returns the cache hit ratio
func (c *Cache) HitRate() float64 {
	stats := c.GetStats()
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0
	}
	return float64(stats.Hits) / float64(total)
}

// ListKeys returns all cached flag keys
func (c *Cache) ListKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0)
	// Ristretto doesn't provide key iteration, so we track separately if needed
	// For now, return empty slice
	return keys
}

// CanEvaluateLocally checks if a flag can be evaluated without Flagr
func (c *Cache) CanEvaluateLocally(flagKey string) bool {
	flag, found := c.GetFlag(flagKey)
	if !found {
		return false
	}
	return c.evaluator.CanEvaluateLocally(flag)
}

// EvaluateLocal performs local evaluation
func (c *Cache) EvaluateLocal(flagKey string, ctx types.EvaluationContext) (bool, error) {
	flag, found := c.GetFlag(flagKey)
	if !found {
		return false, fmt.Errorf("flag not found: %s", flagKey)
	}

	if !c.evaluator.CanEvaluateLocally(flag) {
		return false, fmt.Errorf("flag cannot be evaluated locally: %s", flagKey)
	}

	return c.evaluator.Evaluate(flag, ctx)
}

// Close gracefully shuts down the cache
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store.Close()
	return nil
}

// calculateCost estimates the memory cost of a flag
func (c *Cache) calculateCost(flag *types.Flag) int64 {
	// Rough estimation: base + segments + constraints
	cost := int64(100) // Base overhead
	cost += int64(len(flag.Key))
	cost += int64(len(flag.Description))
	cost += int64(len(flag.Segments) * 200) // Approximate segment size
	return cost
}

// persistFlag saves a single flag to disk
func (c *Cache) persistFlag(flag *types.Flag) error {
	if err := os.MkdirAll(c.persistencePath, 0755); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	data, err := json.Marshal(flag)
	if err != nil {
		return fmt.Errorf("failed to marshal flag: %w", err)
	}

	path := filepath.Join(c.persistencePath, fmt.Sprintf("%s.json", flag.Key))
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write flag to disk: %w", err)
	}

	return nil
}

// persistAllFlags saves multiple flags to disk
func (c *Cache) persistAllFlags(flags []*types.Flag) error {
	if err := os.MkdirAll(c.persistencePath, 0755); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	for _, flag := range flags {
		if flag == nil {
			continue
		}

		data, err := json.Marshal(flag)
		if err != nil {
			continue // Skip failed marshals
		}

		path := filepath.Join(c.persistencePath, fmt.Sprintf("%s.json", flag.Key))
		os.WriteFile(path, data, 0644) // Ignore errors
	}

	return nil
}

// loadFromDisk restores flags from persistent storage
func (c *Cache) loadFromDisk() error {
	if _, err := os.Stat(c.persistencePath); os.IsNotExist(err) {
		return nil // No persisted data yet
	}

	entries, err := os.ReadDir(c.persistencePath)
	if err != nil {
		return fmt.Errorf("failed to read persistence directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(c.persistencePath, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue // Skip failed reads
		}

		var flag types.Flag
		if err := json.Unmarshal(data, &flag); err != nil {
			continue // Skip invalid JSON
		}

		// Restore to cache (without re-persisting)
		cost := c.calculateCost(&flag)
		c.store.Set(flag.Key, &flag, cost)
	}

	return nil
}

// WarmUp preloads flags into cache
func (c *Cache) WarmUp(ctx context.Context, flags []*types.Flag) error {
	return c.SetFlags(flags)
}

// Ensure Cache implements FlagCache interface
var _ types.FlagCache = (*Cache)(nil)
