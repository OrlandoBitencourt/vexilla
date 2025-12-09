package main

import (
	"context"
	"log"
	"time"

	flagrcache "github.com/OrlandoBitencourt/vexilla"
)

func main() {
	// Create configuration
	config := flagrcache.DefaultConfig()
	config.FlagrEndpoint = "localhost:18000"

	// Create cache
	cache, err := flagrcache.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Start background services
	if err := cache.Start(); err != nil {
		log.Fatal(err)
	}
	defer cache.Stop()

	ctx := context.Background()

	// Simple boolean flag
	enabled := cache.EvaluateBool(ctx, "new_feature", flagrcache.EvaluationContext{
		UserID: "user123",
		Attributes: map[string]interface{}{
			"country": "US",
			"age":     25,
		},
	})

	log.Printf("Feature enabled: %v", enabled)

	// String flag with default
	theme := cache.EvaluateString(ctx, "ui_theme", flagrcache.EvaluationContext{
		UserID: "user123",
	}, "light")

	log.Printf("Theme: %s", theme)

	// Keep running
	time.Sleep(10 * time.Minute)
}
