package main

import (
	"context"
	"log"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

func main() {
	// Create configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"

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

	// Example 1: Boolean flag
	log.Println("\n=== Boolean Flag ===")
	enabled := client.EvaluateBool(ctx, "new_feature", vexilla.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "US",
			"tier":    "premium",
		},
	})
	log.Printf("✅ New feature enabled: %v", enabled)

	// Example 2: String variant
	log.Println("\n=== String Variant ===")
	theme := client.EvaluateString(ctx, "ui_theme", vexilla.EvaluationContext{
		EntityID: "user123",
	}, "light")
	log.Printf("🎨 UI Theme: %s", theme)

	// Example 3: Full evaluation result
	log.Println("\n=== Full Evaluation ===")
	result, err := client.Evaluate(ctx, "brazil_launch", vexilla.EvaluationContext{
		EntityID: "user789",
		Context: map[string]interface{}{
			"country":  "BR",
			"document": "1234567808",
		},
	})
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		log.Printf("✅ Variant: %s", result.VariantKey)
		log.Printf("   Evaluated locally: %v", result.EvaluatedLocally)
		log.Printf("   Time: %v", result.EvaluationTime)
	}

	log.Println("\n✨ Demo complete!")
}
