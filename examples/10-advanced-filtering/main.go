// Package main demonstrates advanced flag filtering for memory optimization.
//
// This example shows how to use filtering to reduce memory usage by 50-95%
// in microservice architectures by caching only relevant flags.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🔍 Vexilla Advanced Filtering Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates advanced filtering for memory optimization.")
	fmt.Println()

	// Example 1: Only enabled flags
	fmt.Println("Example 1: Filter Only Enabled Flags")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario: Cache only enabled flags to reduce memory usage")
	fmt.Println()

	fmt.Println("Creating client with OnlyEnabled filter...")
	client1, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true), // Filter out disabled flags
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client1.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started with OnlyEnabled filter")
	fmt.Println()

	metrics1 := client1.Metrics()
	fmt.Printf("Flags cached (enabled only): %d\n", metrics1.Storage.KeysAdded)
	fmt.Println()

	client1.Stop()

	// Example 2: Service-specific filtering
	fmt.Println("Example 2: Service-Specific Tag Filtering")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario: Each microservice caches only its relevant flags")
	fmt.Println()

	services := []struct {
		name string
		tags []string
	}{
		{"user-service", []string{"user", "authentication"}},
		{"payment-service", []string{"payment", "billing"}},
		{"notification-service", []string{"notification", "email"}},
	}

	for _, svc := range services {
		fmt.Printf("Service: %-20s Tags: %v\n", svc.name, svc.tags)
	}
	fmt.Println()

	// Simulate user-service
	fmt.Println("Creating client for 'user-service'...")
	client2, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
		vexilla.WithServiceTag("user-service"), // Only cache flags tagged with service name
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if err := client2.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started with service tag filter")
	fmt.Println()

	metrics2 := client2.Metrics()
	fmt.Printf("Flags cached (user-service only): %d\n", metrics2.Storage.KeysAdded)
	fmt.Println()

	client2.Stop()

	// Example 3: Multiple tag filtering (ANY mode)
	fmt.Println("Example 3: Multiple Tag Filtering (ANY match)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario: Cache flags matching ANY of the specified tags")
	fmt.Println()

	fmt.Println("Creating client with multiple tags (user OR premium)...")
	client3, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
		vexilla.WithAdditionalTags([]string{"user", "premium"}, "any"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if err := client3.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started with ANY tag matching")
	fmt.Println()

	metrics3 := client3.Metrics()
	fmt.Printf("Flags cached (user OR premium): %d\n", metrics3.Storage.KeysAdded)
	fmt.Println()

	client3.Stop()

	// Example 4: Multiple tag filtering (ALL mode)
	fmt.Println("Example 4: Multiple Tag Filtering (ALL match)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario: Cache flags matching ALL of the specified tags")
	fmt.Println()

	fmt.Println("Creating client with multiple tags (user AND premium)...")
	client4, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
		vexilla.WithAdditionalTags([]string{"user", "premium"}, "all"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if err := client4.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started with ALL tag matching")
	fmt.Println()

	metrics4 := client4.Metrics()
	fmt.Printf("Flags cached (user AND premium): %d\n", metrics4.Storage.KeysAdded)
	fmt.Println()

	client4.Stop()

	// Example 5: Combined filtering
	fmt.Println("Example 5: Combined Filtering Strategy")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Scenario: Combine multiple filters for maximum optimization")
	fmt.Println("  • Only enabled flags")
	fmt.Println("  • Service-specific tag")
	fmt.Println("  • Additional environment tags")
	fmt.Println()

	fmt.Println("Creating client with combined filters...")
	client5, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
		vexilla.WithServiceTag("user-service"),
		vexilla.WithAdditionalTags([]string{"production"}, "all"),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if err := client5.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started with combined filters")
	fmt.Println()

	metrics5 := client5.Metrics()
	fmt.Printf("Flags cached (combined filters): %d\n", metrics5.Storage.KeysAdded)
	fmt.Println()

	client5.Stop()

	// Example 6: Memory savings calculation
	fmt.Println("Example 6: Memory Savings Analysis")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Simulate different scenarios
	totalFlags := 10000      // Typical production scenario
	bytesPerFlag := 1024     // ~1KB per flag

	scenarios := []struct {
		name         string
		filteredPct  float64
		description  string
	}{
		{"No filtering", 0, "All flags cached"},
		{"OnlyEnabled", 30, "30% disabled flags removed"},
		{"Service tags", 95, "95% irrelevant flags removed"},
		{"Combined", 97, "97% optimized for microservice"},
	}

	fmt.Println("Memory Usage Comparison:")
	fmt.Println()
	fmt.Printf("Scenario                    | Flags Cached | Memory Used | Saved    | % Saved\n")
	fmt.Println(repeat("-", 80))

	for _, scenario := range scenarios {
		cachedFlags := int(float64(totalFlags) * (1 - scenario.filteredPct/100))
		memoryUsed := cachedFlags * bytesPerFlag
		memoryTotal := totalFlags * bytesPerFlag
		memorySaved := memoryTotal - memoryUsed
		percentSaved := scenario.filteredPct

		fmt.Printf("%-27s | %12d | %8.2f MB | %7.2f MB | %6.1f%%\n",
			scenario.name,
			cachedFlags,
			float64(memoryUsed)/1024/1024,
			float64(memorySaved)/1024/1024,
			percentSaved,
		)
	}

	fmt.Println()

	// Example 7: Real-world microservice example
	fmt.Println("Example 7: Real-World Microservice Architecture")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Architecture: E-commerce platform with 8 microservices")
	fmt.Println()
	fmt.Println("Total flags in Flagr: 10,000")
	fmt.Println("Average flags per service: 500 (5% of total)")
	fmt.Println()

	microservices := []struct {
		name     string
		flags    int
		memory   float64
	}{
		{"api-gateway", 300, 0.29},
		{"user-service", 450, 0.44},
		{"auth-service", 200, 0.20},
		{"product-service", 600, 0.59},
		{"cart-service", 400, 0.39},
		{"payment-service", 350, 0.34},
		{"notification-service", 250, 0.24},
		{"analytics-service", 500, 0.49},
	}

	totalWithoutFiltering := float64(len(microservices) * totalFlags * bytesPerFlag) / 1024 / 1024
	totalWithFiltering := 0.0

	fmt.Println("Per-Service Memory Usage (with filtering):")
	fmt.Println()

	for _, svc := range microservices {
		fmt.Printf("%-25s: %4d flags, %6.2f MB\n", svc.name, svc.flags, svc.memory)
		totalWithFiltering += svc.memory
	}

	fmt.Println()
	fmt.Printf("Total without filtering: %.2f MB (%d services × 10,000 flags)\n",
		totalWithoutFiltering, len(microservices))
	fmt.Printf("Total with filtering:    %.2f MB\n", totalWithFiltering)
	fmt.Printf("Total saved:             %.2f MB (%.1f%% reduction)\n",
		totalWithoutFiltering-totalWithFiltering,
		(totalWithoutFiltering-totalWithFiltering)/totalWithoutFiltering*100,
	)
	fmt.Println()

	// Example 8: Dynamic filtering recommendations
	fmt.Println("Example 8: Filtering Strategy Recommendations")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	strategies := []struct {
		scenario    string
		filters     []string
		savings     string
		recommended bool
	}{
		{
			"Monolithic application",
			[]string{"OnlyEnabled(true)"},
			"20-40%",
			true,
		},
		{
			"Microservice (single service)",
			[]string{"OnlyEnabled(true)", "ServiceTag(serviceName, true)"},
			"90-95%",
			true,
		},
		{
			"Multi-tenant service",
			[]string{"OnlyEnabled(true)", "AdditionalTags([\"tenant-x\"], \"any\")"},
			"50-80%",
			true,
		},
		{
			"Environment-specific",
			[]string{"OnlyEnabled(true)", "AdditionalTags([\"production\"], \"all\")"},
			"30-50%",
			true,
		},
		{
			"No filtering",
			[]string{},
			"0%",
			false,
		},
	}

	for _, strategy := range strategies {
		recommended := "   "
		if strategy.recommended {
			recommended = "✅ "
		} else {
			recommended = "⚠️  "
		}

		fmt.Printf("%s%-30s\n", recommended, strategy.scenario)
		fmt.Printf("   Filters: %v\n", strategy.filters)
		fmt.Printf("   Expected savings: %s\n", strategy.savings)
		fmt.Println()
	}

	fmt.Println(repeat("=", 80))
	fmt.Println("✅ Advanced filtering example completed!")
	fmt.Println()
	fmt.Println("🔑 Key Takeaways:")
	fmt.Println("   ✅ OnlyEnabled(true) - Always use to filter disabled flags")
	fmt.Println("   ✅ ServiceTag() - Essential for microservices (90-95% savings)")
	fmt.Println("   ✅ AdditionalTags() - Fine-tune with environment/feature tags")
	fmt.Println("   ✅ Combined filters - Stack multiple filters for maximum optimization")
	fmt.Println()
	fmt.Println("💡 Best Practices:")
	fmt.Println("   • Tag flags in Flagr with service names")
	fmt.Println("   • Use consistent tagging conventions")
	fmt.Println("   • Monitor cache metrics to verify filtering effectiveness")
	fmt.Println("   • Start conservative, optimize based on metrics")
	fmt.Println()
	fmt.Println("📊 Expected Results:")
	fmt.Println("   • Monolithic: 20-40% memory reduction")
	fmt.Println("   • Microservices: 90-95% memory reduction")
	fmt.Println("   • Multi-tenant: 50-80% memory reduction")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
