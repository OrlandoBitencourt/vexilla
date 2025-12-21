// Package main demonstrates the concept of disk persistence for warm cache on restart.
//
// NOTE: This is currently a conceptual example. Disk persistence is implemented
// internally but not yet exposed in the public API. This example demonstrates
// the benefits and use cases for when this feature becomes available.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("💾 Vexilla Disk Persistence Example (Conceptual)")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates the concept of disk persistence for fast startup.")
	fmt.Println()
	fmt.Println("⚠️  NOTE: Disk persistence is currently implemented internally but not yet")
	fmt.Println("   exposed in the public API. This example shows the benefits and patterns.")
	fmt.Println()

	// Example 1: Current behavior - Cold cache on every start
	fmt.Println("Example 1: Current Behavior (No Disk Persistence)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Creating client...")
	start := time.Now()

	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}
	defer client.Stop()

	startupDuration := time.Since(start)
	fmt.Printf("✅ Client started\n")
	fmt.Printf("   Startup time: %v (cold cache - fetches from Flagr)\n", startupDuration)
	fmt.Println()

	// Perform some evaluations
	fmt.Println("Performing test evaluations...")
	testFlags := []string{
		"new_feature",
		"premium_features",
		"dark_mode",
		"beta_access",
	}

	evalCtx := vexilla.NewContext("persistence-test-user").
		WithAttribute("tier", "premium")

	for _, flagKey := range testFlags {
		enabled := client.Bool(ctx, flagKey, evalCtx)
		status := "❌"
		if enabled {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, flagKey)
	}
	fmt.Println()

	metrics := client.Metrics()
	fmt.Printf("Cache statistics:\n")
	fmt.Printf("  Flags cached: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  Hit ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Println()

	// Example 2: Conceptual disk persistence behavior
	fmt.Println("Example 2: Future Disk Persistence Behavior (Conceptual)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("With disk persistence enabled, the behavior would be:")
	fmt.Println()

	fmt.Println("First startup (cold cache):")
	fmt.Println("  1. Fetch all flags from Flagr (500-2000ms)")
	fmt.Println("  2. Cache in memory (Ristretto)")
	fmt.Println("  3. Save to disk (JSON files)")
	fmt.Println("  4. Ready to serve requests")
	fmt.Println()

	fmt.Println("Subsequent startups (warm cache):")
	fmt.Println("  1. Load flags from disk (50-100ms)")
	fmt.Println("  2. Populate memory cache")
	fmt.Println("  3. Refresh from Flagr in background")
	fmt.Println("  4. Ready immediately!")
	fmt.Println()

	// Example 3: Performance comparison
	fmt.Println("Example 3: Performance Comparison (Projected)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Startup Performance:")
	fmt.Printf("  Cold start (current):    %v\n", startupDuration)
	fmt.Printf("  Warm start (with disk):  ~50-100ms (estimated 10-20x faster)\n")
	fmt.Println()

	// Example 4: Recovery scenarios
	fmt.Println("Example 4: Recovery Scenarios")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario 1: Application Restart")
	fmt.Println("  Current:  Must fetch from Flagr (~1-2s)")
	fmt.Println("  With disk: Load from disk (~50-100ms)")
	fmt.Println()

	fmt.Println("Scenario 2: Flagr Temporarily Down")
	fmt.Println("  Current:  Start fails if Flagr unavailable")
	fmt.Println("  With disk: Use last-known-good state from disk")
	fmt.Println()

	fmt.Println("Scenario 3: Network Issues")
	fmt.Println("  Current:  Cannot start without Flagr connection")
	fmt.Println("  With disk: Graceful degradation using cached state")
	fmt.Println()

	// Example 5: Future API (conceptual)
	fmt.Println("Example 5: Future API Design (Conceptual)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Proposed API for disk persistence:")
	fmt.Println()
	fmt.Println("  client, err := vexilla.New(")
	fmt.Println("      vexilla.WithFlagrEndpoint(\"http://localhost:18000\"),")
	fmt.Println("      vexilla.WithDiskCache(\"/var/cache/myapp/vexilla\"), // Enable disk persistence")
	fmt.Println("      vexilla.WithRefreshInterval(5*time.Minute),")
	fmt.Println("  )")
	fmt.Println()

	fmt.Println("Cache directory structure:")
	fmt.Println("  /var/cache/myapp/vexilla/")
	fmt.Println("    ├── new_feature.json")
	fmt.Println("    ├── premium_features.json")
	fmt.Println("    └── dark_mode.json")
	fmt.Println()

	// Example 6: Benefits summary
	fmt.Println("Example 6: Benefits of Disk Persistence")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("✅ Benefits:")
	fmt.Println("   • Fast startup: 10-20x faster (50-100ms vs 1-2s)")
	fmt.Println("   • Resilience: Works when Flagr is temporarily unavailable")
	fmt.Println("   • Recovery: Last-known-good state preserved across restarts")
	fmt.Println("   • Offline: Can start without network connectivity")
	fmt.Println("   • Zero data loss: Survives application crashes")
	fmt.Println()

	fmt.Println("⚠️  Trade-offs:")
	fmt.Println("   • Disk I/O overhead (mitigated by async writes)")
	fmt.Println("   • Potential stale data (mitigated by TTL + background refresh)")
	fmt.Println("   • Disk space usage (minimal: ~1KB per flag)")
	fmt.Println()

	// Example 7: Memory savings with filtering
	fmt.Println("Example 7: Combined with Filtering for Maximum Efficiency")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	totalFlags := 10000
	serviceFlagsOnly := 500
	bytesPerFlag := 1024

	withoutFiltering := float64(totalFlags * bytesPerFlag / 1024 / 1024)
	withFiltering := float64(serviceFlagsOnly * bytesPerFlag / 1024 / 1024)

	fmt.Printf("Disk usage comparison (10,000 total flags):\n")
	fmt.Printf("  Without filtering: %.2f MB (%d flags)\n", withoutFiltering, totalFlags)
	fmt.Printf("  With filtering:    %.2f MB (%d flags for this service)\n", withFiltering, serviceFlagsOnly)
	fmt.Printf("  Disk space saved:  %.2f MB (%.0f%% reduction)\n",
		withoutFiltering-withFiltering,
		(withoutFiltering-withFiltering)/withoutFiltering*100)
	fmt.Println()

	// Example 8: Use cases
	fmt.Println("Example 8: Ideal Use Cases for Disk Persistence")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	useCases := []struct {
		scenario    string
		benefit     string
		recommended bool
	}{
		{
			"High-availability services",
			"Fast startup for quick recovery",
			true,
		},
		{
			"Edge computing",
			"Offline capability in limited connectivity",
			true,
		},
		{
			"Mobile backends",
			"Resilience during network issues",
			true,
		},
		{
			"Frequent restarts",
			"Avoid repeated Flagr API calls",
			true,
		},
		{
			"Stateless containers",
			"Warm cache on pod restart",
			true,
		},
		{
			"Development environments",
			"Faster dev iteration cycles",
			false,
		},
	}

	for _, uc := range useCases {
		recommended := "⚠️ "
		if uc.recommended {
			recommended = "✅ "
		}

		fmt.Printf("%s%-30s: %s\n", recommended, uc.scenario, uc.benefit)
	}
	fmt.Println()

	fmt.Println(repeat("=", 80))
	fmt.Println("✅ Disk persistence concept example completed!")
	fmt.Println()
	fmt.Println("🔑 Key Insights:")
	fmt.Println("   • Disk persistence enables 10-20x faster startup")
	fmt.Println("   • Provides last-known-good fallback during outages")
	fmt.Println("   • Minimal overhead (~1KB per flag)")
	fmt.Println("   • Ideal for production high-availability scenarios")
	fmt.Println()
	fmt.Println("📝 Current Status:")
	fmt.Println("   • Feature is implemented internally")
	fmt.Println("   • Not yet exposed in public API")
	fmt.Println("   • Coming in future release")
	fmt.Println()
	fmt.Println("💡 Alternative Approaches (Current):")
	fmt.Println("   • Use circuit breaker for resilience (see 08-circuit-breaker)")
	fmt.Println("   • Enable webhook invalidation for real-time updates (see 04-webhook-invalidation)")
	fmt.Println("   • Optimize with filtering to reduce memory/load time (see 10-advanced-filtering)")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
