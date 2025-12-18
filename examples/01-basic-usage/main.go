// Package main demonstrates basic Vexilla usage.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🏴 Vexilla Basic Example")
	fmt.Println(repeat("=", 70))
	fmt.Println()

	// Step 1: Create client
	fmt.Println("📦 Creating Vexilla client...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	// Step 2: Start client (loads flags)
	fmt.Println("🚀 Starting client and loading flags...")
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Client started successfully!")
	fmt.Println()

	// Example 1: Simple boolean flag
	fmt.Println("Example 1: Boolean Flag Evaluation")
	fmt.Println(repeat("-", 70))

	evalCtx := vexilla.NewContext("user-123")

	enabled := client.Bool(ctx, "new_feature", evalCtx)
	fmt.Printf("Flag: new_feature\n")
	fmt.Printf("User: user-123\n")
	fmt.Printf("Result: %v\n", enabled)
	fmt.Println()

	// Example 2: Boolean flag with attributes
	fmt.Println("Example 2: Boolean Flag with Constraints")
	fmt.Println(repeat("-", 70))

	premiumCtx := vexilla.NewContext("user-456").
		WithAttribute("tier", "premium")

	premiumEnabled := client.Bool(ctx, "premium_features", premiumCtx)
	fmt.Printf("Flag: premium_features\n")
	fmt.Printf("User: user-456 (tier=premium)\n")
	fmt.Printf("Result: %v\n", premiumEnabled)
	fmt.Println()

	// Example 3: String flag (A/B test)
	fmt.Println("Example 3: String Flag (UI Theme)")
	fmt.Println(repeat("-", 70))

	themeCtx := vexilla.NewContext("user-789")
	theme := client.String(ctx, "ui_theme", themeCtx, "light")

	fmt.Printf("Flag: ui_theme\n")
	fmt.Printf("User: user-789\n")
	fmt.Printf("Theme: %s\n", theme)
	fmt.Println()

	// Example 4: Integer flag
	fmt.Println("Example 4: Integer Flag (Max Items)")
	fmt.Println(repeat("-", 70))

	maxCtx := vexilla.NewContext("user-999")
	maxItems := client.Int(ctx, "max_items", maxCtx, 10)

	fmt.Printf("Flag: max_items\n")
	fmt.Printf("User: user-999\n")
	fmt.Printf("Max items: %d\n", maxItems)
	fmt.Println()

	// Example 5: Detailed evaluation
	fmt.Println("Example 5: Detailed Evaluation")
	fmt.Println(repeat("-", 70))

	detailCtx := vexilla.NewContext("user-detail")
	result, err := client.Evaluate(ctx, "dark_mode", detailCtx)

	if err != nil {
		fmt.Printf("❌ Evaluation failed: %v\n", err)
	} else {
		fmt.Printf("Flag Key: %s\n", result.FlagKey)
		fmt.Printf("Variant: %s\n", result.VariantKey)
		fmt.Printf("Is Enabled: %v\n", result.IsEnabled())
		fmt.Printf("Reason: %s\n", result.EvaluationReason)

		// Access custom attachment data
		if value := result.GetString("value", ""); value != "" {
			fmt.Printf("Custom Value: %v\n", value)
		}
	}
	fmt.Println()

	// Example 6: Testing different users
	fmt.Println("Example 6: Multiple User Contexts")
	fmt.Println(repeat("-", 70))

	users := []struct {
		id   string
		tier string
	}{
		{"user-001", "free"},
		{"user-002", "premium"},
		{"user-003", "enterprise"},
	}

	for _, u := range users {
		userCtx := vexilla.NewContext(u.id).
			WithAttribute("tier", u.tier)

		enabled := client.Bool(ctx, "premium_features", userCtx)
		fmt.Printf("%-12s (tier=%-10s): %v\n", u.id, u.tier, enabled)
	}
	fmt.Println()

	// Example 7: A/B Testing
	fmt.Println("Example 7: A/B Test Distribution")
	fmt.Println(repeat("-", 70))

	variants := make(map[string]int)

	// Simulate 100 users
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("test-user-%03d", i)
		userCtx := vexilla.NewContext(userID)

		result, err := client.Evaluate(ctx, "button_color_test", userCtx)
		if err == nil {
			variants[result.VariantKey]++
		}
	}

	fmt.Println("Button Color A/B Test Results:")
	for variant, count := range variants {
		fmt.Printf("  %s: %d%% (%d users)\n", variant, count, count)
	}
	fmt.Println()

	// Example 8: Cache metrics
	fmt.Println("Example 8: Performance Metrics")
	fmt.Println(repeat("-", 70))

	metrics := client.Metrics()

	fmt.Println("Cache Performance:")
	fmt.Printf("  Keys Cached: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  Keys Evicted: %d\n", metrics.Storage.KeysEvicted)
	fmt.Printf("  Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)

	fmt.Println("\nHealth Status:")
	fmt.Printf("  Last Refresh: %s ago\n", time.Since(metrics.LastRefresh).Round(time.Second))
	fmt.Printf("  Circuit Open: %v\n", metrics.CircuitOpen)
	fmt.Printf("  Consecutive Fails: %d\n", metrics.ConsecutiveFails)
	fmt.Println()

	fmt.Println(repeat("=", 70))
	fmt.Println("✅ Example completed successfully!")
	fmt.Println()
	fmt.Println("💡 Try these commands:")
	fmt.Println("   • Modify flags in Flagr UI: http://localhost:18000")
	fmt.Println("   • Run this example again to see changes")
	fmt.Println("   • Check internal/cache metrics for performance data")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
