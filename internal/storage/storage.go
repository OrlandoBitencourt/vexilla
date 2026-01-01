// internal/storage/storage.go
package storage

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"

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
// All fields use atomic types for safe concurrent access
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

// atomicMetrics holds the atomic versions of metrics for thread-safe updates
type atomicMetrics struct {
	// Cache statistics
	keysAdded   atomic.Uint64
	keysUpdated atomic.Uint64
	keysEvicted atomic.Uint64
	keysDeleted atomic.Uint64

	// Memory statistics
	costAdded   atomic.Uint64
	costEvicted atomic.Uint64

	// Operation statistics
	setsDropped  atomic.Uint64
	setsRejected atomic.Uint64
	getsKept     atomic.Uint64
	getsDropped  atomic.Uint64

	// Performance metrics
	hitRatio atomic.Uint64 // stored as bits of float64

	// Size
	size atomic.Int64
}

// snapshot returns a consistent snapshot of the metrics
func (am *atomicMetrics) snapshot() Metrics {
	return Metrics{
		KeysAdded:    am.keysAdded.Load(),
		KeysUpdated:  am.keysUpdated.Load(),
		KeysEvicted:  am.keysEvicted.Load(),
		KeysDeleted:  am.keysDeleted.Load(),
		CostAdded:    am.costAdded.Load(),
		CostEvicted:  am.costEvicted.Load(),
		SetsDropped:  am.setsDropped.Load(),
		SetsRejected: am.setsRejected.Load(),
		GetsKept:     am.getsKept.Load(),
		GetsDropped:  am.getsDropped.Load(),
		HitRatio:     float64frombits(am.hitRatio.Load()),
		Size:         am.size.Load(),
	}
}

// float64frombits converts uint64 bits to float64
func float64frombits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// float64bits converts float64 to uint64 bits
func float64bits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
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
