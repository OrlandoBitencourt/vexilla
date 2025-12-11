// examples/basic/main.go
package main

import (
	"context"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	log.Println("🏴 Vexilla - Basic Example")

	// 1. Create configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.RefreshInterval = 1 * time.Minute
	config.PersistenceEnabled = true
	config.PersistencePath = "./vexilla-Cache"

	// 2. Create vexillaCache instance
	log.Println("📦 Creating Vexilla vexillaCache...")
	vexillaCache, err := vexilla.New(config)
	if err != nil {
		log.Fatalf("Failed to create vexillaCache: %v", err)
	}

	// 3. Start the vexillaCache (loads flags and starts background refresh)
	log.Println("🚀 Starting vexillaCache...")
	if err := vexillaCache.Start(); err != nil {
		log.Fatalf("Failed to start vexillaCache: %v", err)
	}
	defer vexillaCache.Stop()

	log.Println("✅ vexillaCache started successfully!")

	// Give it a moment to load flags
	time.Sleep(500 * time.Millisecond)

	ctx := context.Background()

	// 4. Boolean flag evaluation
	log.Println("=== Example 1: Boolean Flag ===")
	evalCtx := vexilla.EvaluationContext{
		UserID: "user_123",
		Attributes: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	enabled := vexillaCache.EvaluateBool(ctx, "new_feature-6a65f91d", evalCtx)
	log.Printf("New feature enabled: %v\n\n", enabled)

	// 5. String flag evaluation
	log.Println("=== Example 2: String Flag ===")
	theme := vexillaCache.EvaluateString(ctx, "ui_theme", evalCtx, "light")
	log.Printf("UI Theme: %s\n\n", theme)

	// 6. Integer flag evaluation
	log.Println("=== Example 3: Integer Flag ===")
	maxItems := vexillaCache.EvaluateInt(ctx, "max_items", evalCtx, 10)
	log.Printf("Max items: %d\n\n", maxItems)

	// 7. Full evaluation with details
	log.Println("=== Example 4: Full Evaluation ===")
	result, err := vexillaCache.Evaluate(ctx, "brazil_launch", vexilla.EvaluationContext{
		UserID: "user_789",
		Attributes: map[string]interface{}{
			"country":  "BR",
			"document": "1234567808",
			"age":      25,
		},
	})
	if err != nil {
		log.Printf("Evaluation error: %v", err)
	} else {
		log.Printf("Result: %v", result)
	}

	// 8. vexillaCache statistics
	log.Println("=== Example 5: vexillaCache Statistics ===")
	stats := vexillaCache.GetCacheStats()
	log.Printf("Keys Added: %d\n", stats.KeysAdded)
	log.Printf("Keys Evicted: %d\n", stats.KeysEvicted)
	log.Printf("Hit Ratio: %.2f%%\n", stats.HitRatio*100)
	log.Printf("Last Refresh: %s\n\n", stats.LastRefresh.Format(time.RFC3339))

	// 9. Manual vexillaCache invalidation
	log.Println("=== Example 6: Manual Invalidation ===")
	log.Println("Invalidating 'new_feature' flag...")
	vexillaCache.InvalidateFlag(ctx, "new_feature")
	log.Println("✅ Flag invalidated")

	// 10. Keep running for a bit to show background refresh
	log.Println("📊 vexillaCache is running. Background refresh will occur every minute.")
	log.Println("Press Ctrl+C to stop.")

	// Simulate some flag evaluations
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 6; i++ {
		<-ticker.C
		enabled := vexillaCache.EvaluateBool(ctx, "new_feature", evalCtx)
		log.Printf("[%s] Feature check: %v", time.Now().Format("15:04:05"), enabled)
	}

	log.Println("\n✨ Example completed!")
}
