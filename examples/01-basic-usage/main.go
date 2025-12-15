// examples/01-basic-usage/main.go
// Basic example showing simple flag evaluation
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("=== Example 1: Basic Flag Usage ===\n")

	// Create a simple cache configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.RefreshInterval = 5 * time.Minute
	config.FilterConfig.OnlyEnabled = true // Only cache enabled flags

	// Initialize cache
	cache, err := vexilla.New(config)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// Start the cache
	if err := cache.Start(); err != nil {
		log.Fatalf("Failed to start cache: %v", err)
	}
	defer cache.Stop()

	// Wait for initial load
	time.Sleep(1 * time.Second)

	ctx := context.Background()

	// Example 1: Boolean flag evaluation
	fmt.Println("1. Boolean Flag Evaluation")
	fmt.Println("   Flag: new_feature")

	evalCtx := vexilla.EvaluationContext{
		UserID: "user_123",
		Attributes: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	enabled := cache.EvaluateBool(ctx, "new_feature", evalCtx)
	fmt.Printf("   Result: %v\n\n", enabled)

	// Example 2: String flag evaluation
	fmt.Println("2. String Flag Evaluation")
	fmt.Println("   Flag: ui_theme")

	theme := cache.EvaluateString(ctx, "ui_theme", evalCtx, "light")
	fmt.Printf("   Theme: %s\n\n", theme)

	// Example 3: Integer flag evaluation
	fmt.Println("3. Integer Flag Evaluation")
	fmt.Println("   Flag: max_items")

	maxItems := cache.EvaluateInt(ctx, "max_items", evalCtx, 10)
	fmt.Printf("   Max Items: %d\n\n", maxItems)

	// Example 4: Full evaluation with details
	fmt.Println("4. Full Evaluation with Details")
	fmt.Println("   Flag: brazil_launch")

	result, err := cache.Evaluate(ctx, "brazil_launch", evalCtx)
	if err != nil {
		fmt.Printf("   Error: %v\n\n", err)
	} else {
		fmt.Printf("   Flag Key: %s\n", result.FlagKey)
		fmt.Printf("   Variant: %s\n", result.VariantKey)
		fmt.Printf("   Reason: %s\n\n", result.EvaluationReason)
	}

	// Example 5: Different user context
	fmt.Println("5. Evaluation with Different User")

	usEvalCtx := vexilla.EvaluationContext{
		UserID: "user_456",
		Attributes: map[string]interface{}{
			"country": "US",
			"tier":    "free",
		},
	}

	usEnabled := cache.EvaluateBool(ctx, "brazil_launch", usEvalCtx)
	fmt.Printf("   US User - Brazil Launch: %v\n\n", usEnabled)

	// Show cache statistics
	fmt.Println("6. Cache Statistics")
	stats := cache.GetCacheStats()
	fmt.Printf("   Flags Loaded: %d\n", stats.KeysAdded)
	fmt.Printf("   Hit Ratio: %.2f%%\n", stats.HitRatio*100)
	fmt.Printf("   Last Refresh: %s\n", stats.LastRefresh.Format(time.RFC3339))

	fmt.Println("\n✅ Example completed successfully!")
}
