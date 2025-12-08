package main

import (
	"context"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

func main() {
	// Advanced configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.RefreshInterval = 2 * time.Minute
	config.PersistenceEnabled = true
	config.PersistencePath = "/tmp/vexilla-demo"

	// Enable webhook for instant updates
	config.WebhookEnabled = true
	config.WebhookPort = 8081
	config.WebhookSecret = "my-secret-key"

	// Enable admin API
	config.AdminAPIEnabled = true
	config.AdminAPIPort = 8082

	log.Println("🏴 Creating Vexilla client with advanced config...")
	client, err := vexilla.New(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🚀 Starting Vexilla with all features...")
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	// Example 1: Static flag (evaluated locally)
	log.Println("\n=== Static Flag Evaluation (Local) ===")
	result, err := client.Evaluate(ctx, "brazil_launch", vexilla.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country":  "BR",
			"document": "1234567808",
		},
	})
	if err == nil {
		log.Printf("✅ Result: %s", result.VariantKey)
		log.Printf("   Evaluated locally: %v", result.EvaluatedLocally)
		log.Printf("   Time: %v", result.EvaluationTime)
		log.Printf("   💡 This was evaluated in <1ms with zero HTTP requests!")
	}

	// Example 2: Dynamic flag (requires Flagr)
	log.Println("\n=== Dynamic Flag Evaluation (Flagr) ===")
	result, err = client.Evaluate(ctx, "ab_test", vexilla.EvaluationContext{
		EntityID: "user456",
		Context: map[string]interface{}{
			"country": "US",
		},
	})
	if err == nil {
		log.Printf("✅ Variant: %s", result.VariantKey)
		log.Printf("   Evaluated locally: %v", result.EvaluatedLocally)
		log.Printf("   Time: %v", result.EvaluationTime)
		log.Printf("   💡 This required Flagr for consistent bucketing")
	}

	// Example 3: Convenience methods
	log.Println("\n=== Convenience Methods ===")
	enabled := client.EvaluateBool(ctx, "premium_feature", vexilla.EvaluationContext{
		EntityID: "user789",
		Context: map[string]interface{}{
			"tier": "premium",
		},
	})
	log.Printf("🎯 Premium feature enabled: %v", enabled)

	// Example 4: Get statistics
	log.Println("\n=== Cache Statistics ===")
	stats, _ := client.GetStats(ctx)
	log.Printf("📊 Stats: %+v", stats)

	// Keep running for webhook/admin testing
	log.Println("\n📡 Services running:")
	log.Printf("   Webhook: http://localhost:%d%s", config.WebhookPort, config.WebhookPath)
	log.Printf("   Admin: http://localhost:%d%s", config.AdminAPIPort, config.AdminAPIPath)
	log.Println("\nTry these commands:")
	log.Printf("  curl http://localhost:%d%s/stats", config.AdminAPIPort, config.AdminAPIPath)
	log.Printf("  curl http://localhost:%d/health", config.AdminAPIPort)
	log.Println("\n⏸️  Press Ctrl+C to stop...")

	select {}
}
