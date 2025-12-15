// internal/storage/storage.go
package storage

import (
	"context"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// Storage defines the interface for flag storage
type Storage interface {
	// Get retrieves a flag by key
	Get(ctx context.Context, key string) (*domain.Flag, error)

	// Set stores a flag with optional TTL
	Set(ctx context.Context, key string, flag domain.Flag, ttl time.Duration) error

	// Delete removes a flag
	Delete(ctx context.Context, key string) error

	// Clear removes all flags
	Clear(ctx context.Context) error

	// List returns all flag keys
	List(ctx context.Context) ([]string, error)

	// Metrics returns storage metrics
	Metrics() Metrics

	// Close closes the storage
	Close() error
}

// Metrics represents storage metrics
type Metrics struct {
	// Cache statistics
	KeysAdded   uint64
	KeysUpdated uint64
	KeysEvicted uint64
	KeysDeleted uint64

	// Memory statistics
	CostAdded   uint64
	CostEvicted uint64

	// Operation statistics
	SetsDropped  uint64
	SetsRejected uint64
	GetsKept     uint64
	GetsDropped  uint64

	// Performance metrics
	HitRatio float64

	// Size
	Size int64
}

// Config holds storage configuration
type Config struct {
	// Memory limits
	MaxCost     int64 // Maximum cache size in bytes
	NumCounters int64 // Number of counters for admission policy
	BufferItems int64 // Number of keys per buffer

	// TTL
	DefaultTTL time.Duration

	// Metrics
	MetricsEnabled bool
}

// DefaultConfig returns default storage configuration
func DefaultConfig() Config {
	return Config{
		MaxCost:        1 << 30, // 1GB
		NumCounters:    1e7,     // 10M counters
		BufferItems:    64,
		DefaultTTL:     5 * time.Minute,
		MetricsEnabled: true,
	}
}
