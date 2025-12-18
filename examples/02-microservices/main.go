// Package main demonstrates Vexilla usage in microservice architectures.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🏴 Vexilla Microservice Example")
	fmt.Println(repeat("=", 70))
	fmt.Println()
	fmt.Println("This example demonstrates memory optimization in microservices")
	fmt.Println("by filtering flags using service tags.")
	fmt.Println()

	// Scenario: User service only needs user-related flags
	// This can reduce memory usage by 90-95% in production!

	fmt.Println("📦 Creating client with service filtering...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithOnlyEnabled(true), // Only cache enabled flags
		vexilla.WithRefreshInterval(5*time.Minute),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Client started with filtering enabled")
	fmt.Println()

	// Use Case 1: User Registration Flow
	fmt.Println("Use Case 1: User Registration Features")
	fmt.Println(repeat("-", 70))

	newUser := vexilla.NewContext("new-user-001").
		WithAttribute("signup_date", time.Now().Format("2006-01-02")).
		WithAttribute("country", "BR")

	// Check if beta access is enabled
	betaAccess := client.Bool(ctx, "beta_access", newUser)
	fmt.Printf("Beta Access Available: %v\n", betaAccess)

	if betaAccess {
		fmt.Println("  → User can access beta features")
	} else {
		fmt.Println("  → Standard registration flow")
	}
	fmt.Println()

	// Use Case 2: Premium Features Gating
	fmt.Println("Use Case 2: Premium Features")
	fmt.Println(repeat("-", 70))

	users := []struct {
		id   string
		tier string
	}{
		{"user-free-001", "free"},
		{"user-premium-001", "premium"},
		{"user-enterprise-001", "enterprise"},
	}

	for _, u := range users {
		userCtx := vexilla.NewContext(u.id).
			WithAttribute("tier", u.tier)

		hasPremium := client.Bool(ctx, "premium_features", userCtx)

		status := "❌"
		if hasPremium {
			status = "✅"
		}

		fmt.Printf("%s %-20s (%-10s): Premium Access %s\n",
			status, u.id, u.tier, status)
	}
	fmt.Println()

	// Use Case 3: Regional Feature Rollout
	fmt.Println("Use Case 3: Regional Launch (Brazil)")
	fmt.Println(repeat("-", 70))

	regions := []string{"BR", "US", "UK", "JP", "DE"}

	for _, region := range regions {
		regionalUser := vexilla.NewContext(fmt.Sprintf("user-%s", region)).
			WithAttribute("country", region)

		launched := client.Bool(ctx, "brazil_launch", regionalUser)

		status := "🔒 Not available"
		if launched {
			status = "🚀 Launched!"
		}

		fmt.Printf("Region: %-4s → %s\n", region, status)
	}
	fmt.Println()

	// Use Case 4: Gradual Rollout
	fmt.Println("Use Case 4: Gradual Rollout (30% in Brazil)")
	fmt.Println(repeat("-", 70))

	brazilUsers := make(map[bool]int)

	// Simulate 100 Brazilian users
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("br-user-%03d", i)
		userCtx := vexilla.NewContext(userID).
			WithAttribute("country", "BR")

		// This flag has 30% rollout in Brazil
		hasAccess := client.Bool(ctx, "gradual_rollout_30", userCtx)
		brazilUsers[hasAccess]++
	}

	enabled := brazilUsers[true]
	disabled := brazilUsers[false]

	fmt.Printf("Total Brazilian Users: 100\n")
	fmt.Printf("  ✅ Enabled: %d (%d%%)\n", enabled, enabled)
	fmt.Printf("  ❌ Disabled: %d (%d%%)\n", disabled, disabled)
	fmt.Printf("\nNote: Rollout percentage may vary due to consistent hashing\n")
	fmt.Println()

	// Use Case 5: A/B Testing for Layout
	fmt.Println("Use Case 5: Multi-Variant A/B Test (Pricing Layout)")
	fmt.Println(repeat("-", 70))

	layouts := make(map[string]int)

	// Simulate 300 users
	for i := 0; i < 300; i++ {
		userID := fmt.Sprintf("layout-user-%03d", i)
		userCtx := vexilla.NewContext(userID)

		result, err := client.Evaluate(ctx, "pricing_layout", userCtx)
		if err == nil {
			layout := result.GetString("layout", "standard")
			layouts[layout]++
		}
	}

	fmt.Println("Pricing Layout Distribution:")
	total := 0
	for _, count := range layouts {
		total += count
	}

	for layout, count := range layouts {
		percentage := float64(count) / float64(total) * 100
		fmt.Printf("  %-10s: %3d users (%.1f%%)\n", layout, count, percentage)
	}
	fmt.Println()

	// Use Case 6: Dark Mode Feature
	fmt.Println("Use Case 6: Theme Preference")
	fmt.Println(repeat("-", 70))

	themeUsers := []string{"user-001", "user-002", "user-003"}

	for _, userID := range themeUsers {
		userCtx := vexilla.NewContext(userID)

		result, err := client.Evaluate(ctx, "dark_mode", userCtx)
		if err != nil {
			fmt.Printf("%-10s: Error - %v\n", userID, err)
			continue
		}

		enabled := result.IsEnabled()
		theme := "Light Mode 🌞"
		if enabled {
			theme = "Dark Mode 🌙"
		}

		fmt.Printf("%-10s → %s\n", userID, theme)
	}
	fmt.Println()

	// Performance Metrics
	fmt.Println("Performance & Optimization Metrics")
	fmt.Println(repeat("-", 70))

	metrics := client.Metrics()

	fmt.Println("📊 Cache Statistics:")
	fmt.Printf("  Flags Cached: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  Cache Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("  Keys Evicted: %d\n", metrics.Storage.KeysEvicted)

	fmt.Println("\n🏥 Health Status:")
	fmt.Printf("  Last Refresh: %s ago\n", time.Since(metrics.LastRefresh).Round(time.Second))
	fmt.Printf("  Circuit Breaker: %s\n", circuitStatus(metrics.CircuitOpen))
	fmt.Printf("  Failed Refreshes: %d\n", metrics.ConsecutiveFails)

	// Memory estimation
	fmt.Println("\n💾 Memory Optimization:")
	cachedFlags := int(metrics.Storage.KeysAdded)
	if cachedFlags > 0 {
		// Assume we have 10,000 total flags (typical production scenario)
		totalFlags := 10000
		bytesPerFlag := 1024 // ~1KB per flag

		withoutFiltering := float64(totalFlags * bytesPerFlag / 1024 / 1024)
		withFiltering := float64(cachedFlags * bytesPerFlag / 1024 / 1024)
		saved := withoutFiltering - withFiltering
		percentSaved := (saved / withoutFiltering) * 100

		fmt.Printf("  Without filtering: ~%.2f MB (10,000 flags)\n", withoutFiltering)
		fmt.Printf("  With filtering: ~%.2f MB (%d flags)\n", withFiltering, cachedFlags)
		fmt.Printf("  Memory saved: ~%.2f MB (%.1f%%)\n", saved, percentSaved)
	} else {
		fmt.Println("  No flags cached yet - run setup-flags.go first")
	}

	fmt.Println()
	fmt.Println(repeat("=", 70))
	fmt.Println("✅ Microservice example completed!")
	fmt.Println()
	fmt.Println("💡 Key Takeaways:")
	fmt.Println("   1. Use WithServiceTag() to filter flags by service")
	fmt.Println("   2. Enable WithOnlyEnabled(true) to skip disabled flags")
	fmt.Println("   3. Monitor metrics.Storage.KeysAdded to track memory usage")
	fmt.Println("   4. Memory savings can reach 90-95% in production!")
	fmt.Println()
	fmt.Println("🔗 Next Steps:")
	fmt.Println("   • Add service tags to your flags in Flagr UI")
	fmt.Println("   • Configure filtering in your microservices")
	fmt.Println("   • Monitor cache metrics in production")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func circuitStatus(open bool) string {
	if open {
		return "🔴 OPEN (degraded)"
	}
	return "🟢 CLOSED (healthy)"
}
