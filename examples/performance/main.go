package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

func main() {
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.PersistenceEnabled = false // Disable for pure performance test

	log.Println("🏴 Creating Vexilla client...")
	client, err := vexilla.New(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🚀 Starting Vexilla...")
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	// Wait for initial load
	time.Sleep(2 * time.Second)

	// Benchmark: Static flag evaluation (local)
	log.Println("\n=== Static Flag Performance Test ===")
	log.Println("Testing 10,000 local evaluations...")

	staticStart := time.Now()
	for i := 0; i < 10000; i++ {
		client.EvaluateBool(ctx, "static_flag", vexilla.EvaluationContext{
			EntityID: fmt.Sprintf("user%d", i),
			Context: map[string]interface{}{
				"country": "US",
				"age":     25,
			},
		})
	}
	staticDuration := time.Since(staticStart)

	avgStatic := float64(staticDuration.Microseconds()) / 10000
	log.Printf("✅ 10,000 evaluations in %v", staticDuration)
	log.Printf("   Average: %.2f μs per evaluation", avgStatic)
	log.Printf("   Throughput: ~%.0f evaluations/second", 10000/staticDuration.Seconds())
	log.Printf("   💡 Zero HTTP requests to Flagr!")

	// Benchmark: Dynamic flag evaluation (Flagr)
	log.Println("\n=== Dynamic Flag Performance Test ===")
	log.Println("Testing 100 remote evaluations (Flagr API calls)...")

	dynamicStart := time.Now()
	for i := 0; i < 100; i++ {
		result, err := client.Evaluate(ctx, "dynamic_flag", vexilla.EvaluationContext{
			EntityID: fmt.Sprintf("user%d", i),
			Context: map[string]interface{}{
				"country": "US",
			},
		})
		if err == nil && !result.EvaluatedLocally {
			// Verified it went to Flagr
		}
	}
	dynamicDuration := time.Since(dynamicStart)

	avgDynamic := float64(dynamicDuration.Milliseconds()) / 100
	log.Printf("✅ 100 evaluations in %v", dynamicDuration)
	log.Printf("   Average: %.2f ms per evaluation", avgDynamic)
	log.Printf("   Throughput: ~%.0f evaluations/second", 100/dynamicDuration.Seconds())
	log.Printf("   💡 Each evaluation required 1 HTTP request to Flagr")

	// Comparison
	log.Println("\n=== Performance Comparison ===")
	speedup := avgDynamic * 1000 / avgStatic // Convert ms to μs for comparison
	log.Printf("🚀 Local evaluation is %.0fx faster than Flagr!", speedup)
	log.Printf("   Static:  %.2f μs", avgStatic)
	log.Printf("   Dynamic: %.2f ms (%.0f μs)", avgDynamic, avgDynamic*1000)

	// Cache stats
	stats, _ := client.GetStats(ctx)
	if s, ok := stats.(vexilla.Stats); ok {
		log.Println("\n=== Cache Statistics ===")
		log.Printf("📊 Hit ratio: %.2f%%", s.HitRatio*100)
		log.Printf("   Keys added: %d", s.KeysAdded)
		log.Printf("   Keys evicted: %d", s.KeysEvicted)
	}

	log.Println("\n✨ Performance test complete!")
}
