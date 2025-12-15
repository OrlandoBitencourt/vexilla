package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

// ExampleTestBasic demonstrates basic Vexilla usage.
func ExampleTestBasic() {
	// Create a new Vexilla client
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start the client
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Evaluate a boolean flag
	evalCtx := vexilla.NewContext("user-123").
		WithAttribute("country", "BR").
		WithAttribute("tier", "premium")

	enabled := client.Bool(ctx, "new_feature", evalCtx)
	fmt.Printf("Feature enabled: %v\n", enabled)

	// Evaluate a string flag
	theme := client.String(ctx, "ui-theme", evalCtx, "light")
	fmt.Printf("Theme: %s\n", theme)

	// Get detailed evaluation result
	result, err := client.Evaluate(ctx, "new-feature", evalCtx)
	if err == nil {
		fmt.Printf("Variant: %s, Reason: %s\n", result.VariantKey, result.EvaluationReason)
	}
}

// ExampleTestMicroservice demonstrates Vexilla usage in a microservice.
func ExampleTestMicroservice() {
	// In a microservice, filter flags by service tag to reduce memory usage
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithServiceTag("user-service"),
		vexilla.WithOnlyEnabled(true),
		vexilla.WithAdditionalTags([]string{"production"}, "any"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// This will only evaluate flags tagged with "user-service"
	evalCtx := vexilla.NewContext("user-123")
	enabled := client.Bool(ctx, "user-service-feature", evalCtx)

	fmt.Printf("Feature enabled: %v\n", enabled)
}

// ExampleTestWithConfig demonstrates using a Config struct.
func ExampleTestWithConfig() {
	cfg := vexilla.DefaultConfig()
	cfg.Flagr.Endpoint = "http://localhost:18000"
	cfg.Cache.RefreshInterval = 10 * time.Minute
	cfg.Cache.Filter.OnlyEnabled = true
	cfg.Cache.Filter.ServiceName = "payment-service"
	cfg.Cache.Filter.RequireServiceTag = true

	client, err := vexilla.New(vexilla.WithConfig(cfg))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Use the client...
}

// ExampleTestmetrics demonstrates accessing cache metrics.
func ExampleTestMetrics() {
	client, _ := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
	)

	ctx := context.Background()
	client.Start(ctx)
	defer client.Stop()

	// Get metrics
	metrics := client.Metrics()

	fmt.Printf("Keys Added: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("Last Refresh: %s\n", metrics.LastRefresh.Format(time.RFC3339))
	fmt.Printf("Circuit Open: %v\n", metrics.CircuitOpen)
}

// ExampleTestComplexEvaluation demonstrates using detailed evaluation results.
func ExampleTestComplexEvaluation() {
	client, _ := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
	)

	ctx := context.Background()
	client.Start(ctx)
	defer client.Stop()

	evalCtx := vexilla.NewContext("user-456").
		WithAttribute("country", "US").
		WithAttribute("age", 25)

	result, err := client.Evaluate(ctx, "ab-test-flag", evalCtx)
	if err != nil {
		log.Printf("Evaluation failed: %v", err)
		return
	}

	// Check if enabled
	if result.IsEnabled() {
		fmt.Println("Feature is enabled")
	}

	// Get custom values from variant attachment
	tier := result.GetString("tier", "free")
	maxItems := result.GetInt("maxTestitems", 10)

	fmt.Printf("Tier: %s, Max Items: %d\n", tier, maxItems)
}

func main() {
	fmt.Println("running vexila basic-usage examples")
	ExampleTestBasic()
	ExampleTestMicroservice()
	ExampleTestWithConfig()
	ExampleTestMetrics()
	ExampleTestComplexEvaluation()
}
