// Package main demonstrates basic Vexilla usage with the new facade API.
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
	fmt.Println("=" + repeat("=", 50))

	// Create a new Vexilla client with simple configuration
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Start the client (loads flags and begins background refresh)
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Client started successfully")
	fmt.Println()

	// Example 1: Boolean flag evaluation
	fmt.Println("Example 1: Boolean Flag")
	fmt.Println("-" + repeat("-", 50))

	evalCtx := vexilla.NewContext("user-123").
		WithAttribute("country", "BR").
		WithAttribute("tier", "premium")

	enabled := client.Bool(ctx, "new-checkout", evalCtx)
	fmt.Printf("Flag: new-checkout\n")
	fmt.Printf("User: user-123 (country=BR, tier=premium)\n")
	fmt.Printf("Result: %v\n", enabled)
	fmt.Println()

	// Example 2: String flag evaluation
	fmt.Println("Example 2: String Flag")
	fmt.Println("-" + repeat("-", 50))

	theme := client.String(ctx, "ui-theme", evalCtx, "light")
	fmt.Printf("Flag: ui-theme\n")
	fmt.Printf("Result: %s\n", theme)
	fmt.Println()

	// Example 3: Integer flag evaluation
	fmt.Println("Example 3: Integer Flag")
	fmt.Println("-" + repeat("-", 50))

	maxItems := client.Int(ctx, "max-items", evalCtx, 10)
	fmt.Printf("Flag: max-items\n")
	fmt.Printf("Result: %d\n", maxItems)
	fmt.Println()

	// Example 4: Detailed evaluation
	fmt.Println("Example 4: Detailed Evaluation")
	fmt.Println("-" + repeat("-", 50))

	result, err := client.Evaluate(ctx, "new-checkout", evalCtx)
	if err != nil {
		fmt.Printf("Evaluation failed: %v\n", err)
	} else {
		fmt.Printf("Flag Key: %s\n", result.FlagKey)
		fmt.Printf("Variant: %s\n", result.VariantKey)
		fmt.Printf("Reason: %s\n", result.EvaluationReason)
		fmt.Printf("Is Enabled: %v\n", result.IsEnabled())
	}
	fmt.Println()

	// Example 5: Different user contexts
	fmt.Println("Example 5: Different User Contexts")
	fmt.Println("-" + repeat("-", 50))

	users := []struct {
		id      string
		country string
		tier    string
	}{
		{"user-001", "BR", "premium"},
		{"user-002", "US", "free"},
		{"user-003", "BR", "free"},
	}

	for _, u := range users {
		ctx := vexilla.NewContext(u.id).
			WithAttribute("country", u.country).
			WithAttribute("tier", u.tier)

		enabled := client.Bool(context.Background(), "new-checkout", ctx)
		fmt.Printf("%s (country=%s, tier=%s): %v\n", u.id, u.country, u.tier, enabled)
	}
	fmt.Println()

	// Example 6: Cache metrics
	fmt.Println("Example 6: Cache Metrics")
	fmt.Println("-" + repeat("-", 50))

	metrics := client.Metrics()
	fmt.Printf("Keys Added: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("Keys Evicted: %d\n", metrics.Storage.KeysEvicted)
	fmt.Printf("Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("Last Refresh: %s\n", metrics.LastRefresh.Format(time.RFC3339))
	fmt.Printf("Circuit Open: %v\n", metrics.CircuitOpen)
	fmt.Printf("Consecutive Fails: %d\n", metrics.ConsecutiveFails)

	fmt.Println()
	fmt.Println("✅ Example completed successfully!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
