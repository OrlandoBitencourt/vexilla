package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// ExampleMemoryStorage_concurrentSafe demonstrates that the storage is safe for concurrent use
func ExampleMemoryStorage_concurrentSafe() {
	// Create a new memory storage
	cfg := Config{
		MaxCost:     10 * 1024 * 1024,
		NumCounters: 10000,
		BufferItems: 64,
		DefaultTTL:  time.Minute,
	}

	storage, err := NewMemoryStorage(cfg)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Create a flag
	flag := domain.Flag{
		Key:     "feature-flag",
		Enabled: true,
	}

	// Launch 100 concurrent goroutines
	const numWorkers = 100
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine performs multiple operations
			for j := 0; j < 10; j++ {
				// Set the flag
				_ = storage.Set(ctx, "feature-flag", flag, time.Minute)

				// Get the flag
				_, _ = storage.Get(ctx, "feature-flag")

				// Try to get a missing flag
				_, _ = storage.Get(ctx, "missing-flag")
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Read final metrics - this is safe even while operations are happening
	metrics := storage.Metrics()

	fmt.Printf("Operations completed successfully\n")
	fmt.Printf("Keys added: %d\n", metrics.KeysAdded)
	fmt.Printf("Successful gets: %d\n", metrics.GetsKept)
	fmt.Printf("Failed gets: %d\n", metrics.GetsDropped)

	// Output:
	// Operations completed successfully
	// Keys added: 1000
	// Successful gets: 1000
	// Failed gets: 1000
}

// ExampleMemoryStorage_metricsSnapshot demonstrates that metrics snapshots are consistent
func ExampleMemoryStorage_metricsSnapshot() {
	cfg := Config{
		MaxCost:     10 * 1024 * 1024,
		NumCounters: 10000,
		BufferItems: 64,
		DefaultTTL:  time.Minute,
	}

	storage, err := NewMemoryStorage(cfg)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	ctx := context.Background()

	// Perform some operations
	flag := domain.Flag{Key: "test", Enabled: true}
	_ = storage.Set(ctx, "test", flag, time.Minute)
	_, _ = storage.Get(ctx, "test")
	_, _ = storage.Get(ctx, "missing")

	// Get a snapshot of metrics
	snapshot := storage.Metrics()

	// The snapshot is a consistent view at a point in time
	fmt.Printf("Snapshot captured\n")
	fmt.Printf("Keys added: %d\n", snapshot.KeysAdded)
	fmt.Printf("Gets kept: %d\n", snapshot.GetsKept)
	fmt.Printf("Gets dropped: %d\n", snapshot.GetsDropped)

	// Output:
	// Snapshot captured
	// Keys added: 1
	// Gets kept: 1
	// Gets dropped: 1
}
