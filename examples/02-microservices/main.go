// Package main demonstrates Vexilla usage in a microservice architecture.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🏴 Vexilla Microservice Example")
	fmt.Println("=" + repeat("=", 60))
	fmt.Println()

	// Scenario: A user-service that only needs flags tagged with "user-service"
	// This dramatically reduces memory usage in microservice environments
	//
	// If you have 10,000 total flags but only 50 are tagged for user-service:
	// - Without filtering: ~10 MB memory
	// - With filtering: ~500 KB memory (95% reduction!)

	fmt.Println("Creating client with service-specific filtering...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithServiceTag("user-service"),                    // Only cache flags tagged "user-service"
		vexilla.WithOnlyEnabled(true),                             // Only cache enabled flags
		vexilla.WithAdditionalTags([]string{"production"}, "any"), // Only production flags
		vexilla.WithRefreshInterval(5*time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Client started with filtering enabled")
	fmt.Println()

	// Example 1: User Registration Flow
	fmt.Println("Example 1: User Registration Flow")
	fmt.Println("-" + repeat("-", 60))

	userCtx := vexilla.NewContext("new-user-456").
		WithAttribute("country", "BR").
		WithAttribute("signup_date", time.Now().Format("2006-01-02"))

	// Check if email verification is required
	emailVerificationRequired := client.Bool(ctx, "require-email-verification", userCtx)
	fmt.Printf("Email verification required: %v\n", emailVerificationRequired)

	// Check if social login is enabled
	socialLoginEnabled := client.Bool(ctx, "social-login-enabled", userCtx)
	fmt.Printf("Social login enabled: %v\n", socialLoginEnabled)

	// Get maximum login attempts
	maxLoginAttempts := client.Int(ctx, "max-login-attempts", userCtx, 3)
	fmt.Printf("Maximum login attempts: %d\n", maxLoginAttempts)

	fmt.Println()

	// Example 2: User Profile Features
	fmt.Println("Example 2: User Profile Features")
	fmt.Println("-" + repeat("-", 60))

	profileCtx := vexilla.NewContext("user-789").
		WithAttribute("country", "BR").
		WithAttribute("account_age_days", 45).
		WithAttribute("tier", "premium")

	// Check avatar upload feature
	avatarUpload := client.Bool(ctx, "avatar-upload-enabled", profileCtx)
	fmt.Printf("Avatar upload enabled: %v\n", avatarUpload)

	// Get profile customization level
	customizationLevel := client.String(ctx, "profile-customization", profileCtx, "basic")
	fmt.Printf("Customization level: %s\n", customizationLevel)

	fmt.Println()

	// Example 3: Regional Feature Rollout
	fmt.Println("Example 3: Regional Feature Rollout")
	fmt.Println("-" + repeat("-", 60))

	regions := []string{"BR", "US", "UK", "JP", "IN"}
	for _, region := range regions {
		ctx := vexilla.NewContext("user-regional").
			WithAttribute("country", region)

		newProfileEnabled := client.Bool(context.Background(), "new-profile-page", ctx)
		fmt.Printf("%s: %v\n", region, newProfileEnabled)
	}

	fmt.Println()

	// Example 4: A/B Test Evaluation
	fmt.Println("Example 4: A/B Test Evaluation")
	fmt.Println("-" + repeat("-", 60))

	abTestCtx := vexilla.NewContext("test-user-001").
		WithAttribute("country", "BR")

	result, err := client.Evaluate(ctx, "onboarding-ab-test", abTestCtx)
	if err != nil {
		fmt.Printf("A/B test evaluation failed: %v\n", err)
	} else {
		fmt.Printf("A/B Test Variant: %s\n", result.VariantKey)

		// Get experiment-specific configuration
		flowType := result.GetString("flow_type", "standard")
		stepCount := result.GetInt("step_count", 3)

		fmt.Printf("Flow Type: %s\n", flowType)
		fmt.Printf("Step Count: %d\n", stepCount)
	}

	fmt.Println()

	// Example 5: Performance Metrics
	fmt.Println("Example 5: Performance & Memory Savings")
	fmt.Println("-" + repeat("-", 60))

	metrics := client.Metrics()

	fmt.Printf("Cache Performance:\n")
	fmt.Printf("  Keys Cached: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("  Last Refresh: %s ago\n", time.Since(metrics.LastRefresh).Round(time.Second))

	fmt.Printf("\nHealth Status:\n")
	fmt.Printf("  Circuit Open: %v\n", metrics.CircuitOpen)
	fmt.Printf("  Consecutive Fails: %d\n", metrics.ConsecutiveFails)

	// Estimate memory savings
	estimatedTotalFlags := 10000
	estimatedCachedFlags := int(metrics.Storage.KeysAdded)
	if estimatedCachedFlags > 0 {
		savedFlags := estimatedTotalFlags - estimatedCachedFlags
		percentSaved := float64(savedFlags) / float64(estimatedTotalFlags) * 100
		memorySavedMB := float64(savedFlags) * 1.0 / 1024 // Assuming ~1KB per flag

		fmt.Printf("\nMemory Optimization (estimated):\n")
		fmt.Printf("  Total flags: %d\n", estimatedTotalFlags)
		fmt.Printf("  Cached flags: %d\n", estimatedCachedFlags)
		fmt.Printf("  Filtered out: %d (%.1f%%)\n", savedFlags, percentSaved)
		fmt.Printf("  Memory saved: ~%.2f MB\n", memorySavedMB)
	}

	fmt.Println()
	fmt.Println("✅ Microservice example completed!")
	fmt.Println()
	fmt.Println("💡 Key Takeaway:")
	fmt.Println("   By filtering flags with service tags, you can reduce memory")
	fmt.Println("   usage by 90-95% in microservice architectures!")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
