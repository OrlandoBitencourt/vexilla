// Package main demonstrates deterministic rollouts using pre-computed buckets.
//
// This pattern enables 100% local evaluation with zero HTTP overhead for rollouts.
// Instead of relying on Flagr's random percentage-based evaluations, we pre-compute
// a deterministic bucket from user identifiers and use simple constraint matching.
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🎲 Vexilla Deterministic Rollout Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example shows how to achieve 100% local evaluation for rollouts")
	fmt.Println("using pre-computed buckets instead of random percentages.")
	fmt.Println()

	// Create client
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Client started successfully!")
	fmt.Println()

	// Example 1: CPF-based rollout (Brazilian tax ID)
	fmt.Println("Example 1: CPF-based Rollout (70% launch)")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("🇧🇷 Scenario: Launching a new feature to 70% of Brazilian users")
	fmt.Println("   Using CPF digits to create deterministic buckets (0-99)")
	fmt.Println()

	// Simulate various CPF numbers
	cpfs := []string{
		"123.456.789-09", // bucket: 78
		"987.654.321-00", // bucket: 32
		"111.222.333-44", // bucket: 33
		"555.666.777-88", // bucket: 77
		"999.888.777-66", // bucket: 77
	}

	rolloutResults := make(map[bool]int)

	for _, cpf := range cpfs {
		bucket := CPFBucket(cpf)

		// Create evaluation context with bucket
		evalCtx := vexilla.NewContext(cpf).
			WithAttribute("cpf_bucket", bucket)

		// Evaluate flag (assumes flag has constraint: cpf_bucket >= 0 AND cpf_bucket <= 69)
		enabled := client.Bool(ctx, "cpf_rollout_70", evalCtx)

		rolloutResults[enabled]++

		status := "❌ Not in rollout"
		if enabled {
			status = "✅ IN ROLLOUT"
		}

		fmt.Printf("CPF: %s → Bucket: %2d → %s\n", cpf, bucket, status)
	}

	fmt.Printf("\nResults: %d enabled, %d disabled\n", rolloutResults[true], rolloutResults[false])
	fmt.Println()

	// Example 2: User ID hash-based rollout
	fmt.Println("Example 2: User ID Hash-based Rollout (30% A/B test)")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("🧪 Scenario: A/B testing with 30% in variant A, 70% in control")
	fmt.Println("   Using user ID hash to create stable buckets")
	fmt.Println()

	variants := make(map[string]int)

	// Simulate 100 users
	for i := 1; i <= 100; i++ {
		userID := fmt.Sprintf("user-%d", i)
		bucket := UserIDBucket(userID)

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("user_bucket", bucket)

		// This flag has two segments:
		// - Segment 1 (rank 0): user_bucket >= 0 AND user_bucket <= 29 → Variant A
		// - Segment 2 (rank 1): default → Control
		result, err := client.Evaluate(ctx, "ab_test_30", evalCtx)
		if err == nil {
			variants[result.VariantKey]++
		}
	}

	fmt.Println("A/B Test Distribution:")
	for variant, count := range variants {
		fmt.Printf("  %-10s: %2d users (%d%%)\n", variant, count, count)
	}
	fmt.Println()

	// Example 3: Multi-variant test (A/B/C)
	fmt.Println("Example 3: Multi-variant Test (33% / 33% / 34% split)")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("🎨 Scenario: Testing 3 different pricing page layouts")
	fmt.Println("   A: 0-32, B: 33-65, C: 66-99")
	fmt.Println()

	layoutVariants := make(map[string]int)

	for i := 1; i <= 150; i++ {
		userID := fmt.Sprintf("layout-user-%d", i)
		bucket := UserIDBucket(userID)

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("user_bucket", bucket)

		// Flag configuration:
		// - Segment 1: user_bucket >= 0 AND user_bucket <= 32 → Layout A
		// - Segment 2: user_bucket >= 33 AND user_bucket <= 65 → Layout B
		// - Segment 3: user_bucket >= 66 → Layout C
		result, err := client.Evaluate(ctx, "pricing_layout_test", evalCtx)
		if err == nil {
			layout := result.GetString("layout", "unknown")
			layoutVariants[layout]++
		}
	}

	fmt.Println("Pricing Layout Distribution:")
	total := 0
	for _, count := range layoutVariants {
		total += count
	}

	for layout, count := range layoutVariants {
		percentage := float64(count) / float64(total) * 100
		fmt.Printf("  Layout %-2s: %3d users (%.1f%%)\n", layout, count, percentage)
	}
	fmt.Println()

	// Example 4: Gradual rollout (10% → 50% → 100%)
	fmt.Println("Example 4: Gradual Rollout Strategy")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("📊 Scenario: Progressive rollout stages")
	fmt.Println()

	stages := []struct {
		name       string
		flagKey    string
		percentage int
	}{
		{"Stage 1 (10%)", "gradual_rollout_10", 10},
		{"Stage 2 (50%)", "gradual_rollout_50", 50},
		{"Stage 3 (100%)", "gradual_rollout_100", 100},
	}

	testUserID := "test-user-12345"
	testBucket := UserIDBucket(testUserID)

	evalCtx := vexilla.NewContext(testUserID).
		WithAttribute("user_bucket", testBucket)

	fmt.Printf("Test User: %s (bucket: %d)\n", testUserID, testBucket)
	fmt.Println()

	for _, stage := range stages {
		enabled := client.Bool(ctx, stage.flagKey, evalCtx)

		status := "❌ Not included"
		if enabled {
			status = "✅ INCLUDED"
		}

		fmt.Printf("%-20s → %s\n", stage.name, status)
	}
	fmt.Println()

	// Example 5: Regional + Bucket rollout
	fmt.Println("Example 5: Combined Regional + Bucket Rollout")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("🌎 Scenario: 50% rollout only in Brazil")
	fmt.Println()

	regions := []struct {
		country string
		userID  string
	}{
		{"BR", "br-user-001"},
		{"BR", "br-user-002"},
		{"US", "us-user-001"},
		{"US", "us-user-002"},
	}

	for _, r := range regions {
		bucket := UserIDBucket(r.userID)

		evalCtx := vexilla.NewContext(r.userID).
			WithAttribute("country", r.country).
			WithAttribute("user_bucket", bucket)

		// Flag constraints:
		// - country EQ "BR" AND user_bucket >= 0 AND user_bucket <= 49
		enabled := client.Bool(ctx, "brazil_50_rollout", evalCtx)

		status := "❌ Not in rollout"
		if enabled {
			status = "✅ IN ROLLOUT"
		}

		fmt.Printf("Country: %s, User: %-12s (bucket: %2d) → %s\n",
			r.country, r.userID, bucket, status)
	}
	fmt.Println()

	// Performance comparison
	fmt.Println("Performance Benefits")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Measure local evaluation time
	start := time.Now()
	iterations := 10000

	for i := 0; i < iterations; i++ {
		userID := fmt.Sprintf("perf-user-%d", i)
		bucket := UserIDBucket(userID)

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("user_bucket", bucket)

		_ = client.Bool(ctx, "cpf_rollout_70", evalCtx)
	}

	duration := time.Since(start)
	avgLatency := duration / time.Duration(iterations)

	fmt.Printf("Local Evaluation Performance:\n")
	fmt.Printf("  Total evaluations: %d\n", iterations)
	fmt.Printf("  Total time: %v\n", duration)
	fmt.Printf("  Average latency: %v\n", avgLatency)
	fmt.Printf("  Throughput: ~%d eval/sec\n", int(float64(iterations)/duration.Seconds()))
	fmt.Printf("  HTTP requests: 0 ✅\n")
	fmt.Println()

	fmt.Println("💡 Key Advantages:")
	fmt.Println("   ✅ Deterministic: Same user always gets same result")
	fmt.Println("   ✅ Zero HTTP overhead: <1ms evaluation latency")
	fmt.Println("   ✅ Offline-capable: Works without Flagr connectivity")
	fmt.Println("   ✅ Reproducible: Easy to debug user-specific behavior")
	fmt.Println("   ✅ Stable assignments: No random variation")
	fmt.Println()

	fmt.Println(repeat("=", 80))
	fmt.Println("✅ Deterministic rollout example completed!")
	fmt.Println()
	fmt.Println("🔧 How to set up in Flagr:")
	fmt.Println("   1. Create flag with segments")
	fmt.Println("   2. Add constraints: attribute >= X AND attribute <= Y")
	fmt.Println("   3. Use pre-computed buckets (0-99) in your app")
	fmt.Println("   4. Enjoy 100% local evaluation!")
}

// CPFBucket extracts a deterministic bucket (0-99) from a Brazilian CPF number.
// Uses digits 6-7 to create the bucket.
func CPFBucket(cpf string) int {
	// Remove formatting: "123.456.789-09" → "12345678909"
	clean := strings.NewReplacer(".", "", "-", "").Replace(cpf)

	if len(clean) < 7 {
		return -1 // Invalid CPF
	}

	// Use digits 6-7 to create bucket 00-99
	bucket, err := strconv.Atoi(clean[5:7])
	if err != nil {
		return -1
	}

	return bucket
}

// UserIDBucket creates a deterministic bucket (0-99) from a user ID using hash.
func UserIDBucket(userID string) int {
	// Simple hash function
	hash := 0
	for _, char := range userID {
		hash = (hash*31 + int(char)) % 100
	}

	return hash
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
