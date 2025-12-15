// examples/01-basic-usage/main.go
// Basic example showing simple flag evaluation
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/cache"
	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/flagr"
	"github.com/OrlandoBitencourt/vexilla/pkg/storage"
)

func main() {
	fmt.Println("=== Example 1: Basic Flag Usage ===\n")

	// Create Flagr client
	flagrClient := flagr.NewHTTPClient(flagr.Config{
		Endpoint:   "http://localhost:18000",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})

	// Create memory storage
	memStorage, err := storage.NewMemoryStorage(storage.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// Create evaluator
	eval := evaluator.New()

	// Create cache with options
	c, err := cache.New(
		cache.WithFlagrClient(flagrClient),
		cache.WithStorage(memStorage),
		cache.WithEvaluator(eval),
		cache.WithRefreshInterval(5*time.Minute),
		cache.WithOnlyEnabled(true), // Only cache enabled flags
	)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// Start the cache
	ctx := context.Background()
	if err := c.Start(ctx); err != nil {
		log.Fatalf("Failed to start cache: %v", err)
	}
	defer c.Stop()

	// Wait for initial load
	time.Sleep(1 * time.Second)

	// Example 1: Boolean flag evaluation
	fmt.Println("1. Boolean Flag Evaluation")
	fmt.Println("   Flag: new_feature")

	evalCtx := domain.EvaluationContext{
		EntityID: "user_123",
		Context: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	enabled := c.EvaluateBool(ctx, "new_feature", evalCtx)
	fmt.Printf("   Result: %v\n\n", enabled)

	// Example 2: String flag evaluation
	fmt.Println("2. String Flag Evaluation")
	fmt.Println("   Flag: ui_theme")

	theme := c.EvaluateString(ctx, "ui_theme", evalCtx, "light")
	fmt.Printf("   Theme: %s\n\n", theme)

	// Example 3: Integer flag evaluation
	fmt.Println("3. Integer Flag Evaluation")
	fmt.Println("   Flag: max_items")

	maxItems := c.EvaluateInt(ctx, "max_items", evalCtx, 10)
	fmt.Printf("   Max Items: %d\n\n", maxItems)

	// Example 4: Full evaluation with details
	fmt.Println("4. Full Evaluation with Details")
	fmt.Println("   Flag: brazil_launch")

	result, err := c.Evaluate(ctx, "brazil_launch", evalCtx)
	if err != nil {
		fmt.Printf("   Error: %v\n\n", err)
	} else {
		fmt.Printf("   Flag Key: %s\n", result.FlagKey)
		fmt.Printf("   Variant: %s\n", result.VariantKey)
		fmt.Printf("   Reason: %s\n\n", result.EvaluationReason)
	}

	// Example 5: Different user context
	fmt.Println("5. Evaluation with Different User")

	usEvalCtx := domain.EvaluationContext{
		EntityID: "user_456",
		Context: map[string]interface{}{
			"country": "US",
			"tier":    "free",
		},
	}

	usEnabled := c.EvaluateBool(ctx, "brazil_launch", usEvalCtx)
	fmt.Printf("   US User - Brazil Launch: %v\n\n", usEnabled)

	// Show cache statistics
	fmt.Println("6. Cache Statistics")
	metrics := c.GetMetrics()
	fmt.Printf("   Storage Metrics:\n")
	fmt.Printf("     Keys Added: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("     Gets Kept: %d\n", metrics.Storage.GetsKept)
	fmt.Printf("     Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("   Last Refresh: %s\n", metrics.LastRefresh.Format(time.RFC3339))

	fmt.Println("\n✅ Example completed successfully!")
}
