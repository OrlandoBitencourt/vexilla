// Package main demonstrates circuit breaker functionality for resilience.
//
// The circuit breaker prevents cascade failures by failing fast when Flagr
// is unavailable, protecting both the application and Flagr infrastructure.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🔌 Vexilla Circuit Breaker Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates circuit breaker protection against failures.")
	fmt.Println()

	// Example 1: Normal operation (Circuit CLOSED)
	fmt.Println("Example 1: Normal Operation (Circuit CLOSED)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Creating client with healthy Flagr endpoint...")
	client1, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(30*time.Second),
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client1.Start(ctx); err != nil {
		log.Fatalf("❌ Failed to start client: %v", err)
	}

	fmt.Println("✅ Client started successfully")
	fmt.Println()

	// Check circuit state
	metrics := client1.Metrics()
	fmt.Printf("Circuit breaker state:\n")
	fmt.Printf("  Status: %s\n", formatCircuitState(metrics.CircuitOpen))
	fmt.Printf("  Consecutive failures: %d\n", metrics.ConsecutiveFails)
	fmt.Printf("  Last refresh: %v ago\n", time.Since(metrics.LastRefresh).Round(time.Second))
	fmt.Println()

	// Perform normal evaluations
	fmt.Println("Performing normal evaluations...")
	evalCtx := vexilla.NewContext("circuit-test-user")

	testFlags := []string{
		"new_feature",
		"premium_features",
		"dark_mode",
	}

	for _, flagKey := range testFlags {
		enabled := client1.Bool(ctx, flagKey, evalCtx)
		status := "❌"
		if enabled {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, flagKey)
	}

	fmt.Println()
	fmt.Println("✅ All evaluations successful (circuit CLOSED)")
	fmt.Println()

	client1.Stop()

	// Example 2: Simulating failures (Circuit OPEN)
	fmt.Println("Example 2: Simulating Failures (Circuit OPEN)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Creating client with INVALID Flagr endpoint to trigger failures...")
	client2, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:99999"), // Invalid port
		vexilla.WithRefreshInterval(5*time.Second), // Short interval for demo
		vexilla.WithOnlyEnabled(true),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	// This will fail to connect to Flagr
	fmt.Println("Attempting to start client (will fail)...")
	if err := client2.Start(ctx); err != nil {
		fmt.Printf("⚠️  Initial connection failed (expected): %v\n", err)
		fmt.Println()
	}

	fmt.Println("Monitoring circuit breaker state changes...")
	fmt.Println()

	// Monitor circuit state over time
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	iterations := 10
	for i := 0; i < iterations; i++ {
		<-ticker.C

		metrics := client2.Metrics()

		timestamp := time.Now().Format("15:04:05")
		state := formatCircuitState(metrics.CircuitOpen)
		fails := metrics.ConsecutiveFails

		fmt.Printf("[%s] Circuit: %-15s | Failures: %d | Last refresh attempt: %v ago\n",
			timestamp,
			state,
			fails,
			time.Since(metrics.LastRefresh).Round(time.Second),
		)

		// Demonstrate evaluation behavior with open circuit
		if i == 5 {
			fmt.Println()
			fmt.Println("  → Attempting flag evaluation with OPEN circuit...")

			enabled := client2.Bool(ctx, "new_feature", evalCtx)
			fmt.Printf("     Result: %v (should be false or default)\n", enabled)

			fmt.Println("     ⚡ Evaluation failed fast (no waiting for timeout)")
			fmt.Println()
		}
	}

	fmt.Println()
	client2.Stop()

	// Example 3: Circuit recovery (HALF-OPEN → CLOSED)
	fmt.Println("Example 3: Circuit Recovery Demonstration")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("This example shows how the circuit recovers when Flagr comes back online.")
	fmt.Println()

	fmt.Println("⚠️  Manual Test Required:")
	fmt.Println()
	fmt.Println("1. Stop Flagr server:")
	fmt.Println("   docker stop flagr")
	fmt.Println()
	fmt.Println("2. Run this example (circuit will OPEN)")
	fmt.Println()
	fmt.Println("3. Start Flagr server:")
	fmt.Println("   docker start flagr")
	fmt.Println()
	fmt.Println("4. Watch circuit transition: OPEN → HALF-OPEN → CLOSED")
	fmt.Println()

	// Example 4: Benefits demonstration
	fmt.Println("Example 4: Benefits of Circuit Breaker")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Without Circuit Breaker:")
	fmt.Println("  ❌ Every evaluation waits for timeout (e.g., 30 seconds)")
	fmt.Println("  ❌ Cascading failures across application")
	fmt.Println("  ❌ High latency even when Flagr is clearly down")
	fmt.Println("  ❌ Resource exhaustion (connections, threads)")
	fmt.Println()

	fmt.Println("With Circuit Breaker:")
	fmt.Println("  ✅ Fail fast (<1ms) when circuit is open")
	fmt.Println("  ✅ Prevents cascade failures")
	fmt.Println("  ✅ Automatic recovery testing (half-open)")
	fmt.Println("  ✅ Protects both application and Flagr")
	fmt.Println()

	// Example 5: Configuration tuning
	fmt.Println("Example 5: Circuit Breaker Configuration")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Default Configuration:")
	fmt.Println("  • Max failures: 3")
	fmt.Println("  • Timeout: 30 seconds")
	fmt.Println("  • Half-open timeout: 10 seconds")
	fmt.Println()

	fmt.Println("Tuning Guidelines:")
	fmt.Println()

	fmt.Println("For high-traffic services:")
	fmt.Println("  • Lower max failures (e.g., 2)")
	fmt.Println("  • Shorter timeout (e.g., 15s)")
	fmt.Println("  • Faster recovery testing")
	fmt.Println()

	fmt.Println("For low-traffic services:")
	fmt.Println("  • Higher max failures (e.g., 5)")
	fmt.Println("  • Longer timeout (e.g., 60s)")
	fmt.Println("  • More conservative recovery")
	fmt.Println()

	// Example 6: State machine
	fmt.Println("Example 6: Circuit Breaker State Machine")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("State Transitions:")
	fmt.Println()
	fmt.Println("  ┌─────────┐")
	fmt.Println("  │ CLOSED  │ ← Normal operation, all requests pass")
	fmt.Println("  └────┬────┘")
	fmt.Println("       │")
	fmt.Println("       │ failures >= max_failures")
	fmt.Println("       ▼")
	fmt.Println("  ┌─────────┐")
	fmt.Println("  │  OPEN   │ ← Blocking requests, failing fast")
	fmt.Println("  └────┬────┘")
	fmt.Println("       │")
	fmt.Println("       │ timeout elapsed")
	fmt.Println("       ▼")
	fmt.Println("  ┌──────────┐")
	fmt.Println("  │ HALF-OPEN│ ← Testing recovery (limited requests)")
	fmt.Println("  └────┬─────┘")
	fmt.Println("       │")
	fmt.Println("       ├─> Success → CLOSED")
	fmt.Println("       └─> Failure → OPEN")
	fmt.Println()

	// Example 7: Monitoring recommendations
	fmt.Println("Example 7: Monitoring Recommendations")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Key Metrics to Monitor:")
	fmt.Println()
	fmt.Println("1. Circuit State:")
	fmt.Println("   • Alert when circuit opens")
	fmt.Println("   • Track open duration")
	fmt.Println("   • Monitor recovery time")
	fmt.Println()

	fmt.Println("2. Failure Count:")
	fmt.Println("   • Consecutive failures")
	fmt.Println("   • Total failures over time")
	fmt.Println("   • Failure rate")
	fmt.Println()

	fmt.Println("3. Recovery Metrics:")
	fmt.Println("   • Time to recovery")
	fmt.Println("   • Recovery success rate")
	fmt.Println("   • Half-open duration")
	fmt.Println()

	fmt.Println("4. Business Impact:")
	fmt.Println("   • Requests affected")
	fmt.Println("   • Fallback usage")
	fmt.Println("   • User experience degradation")
	fmt.Println()

	fmt.Println(repeat("=", 80))
	fmt.Println("✅ Circuit breaker example completed!")
	fmt.Println()
	fmt.Println("🔑 Key Takeaways:")
	fmt.Println("   ✅ Circuit breaker prevents cascade failures")
	fmt.Println("   ✅ Fail fast when Flagr is down (<1ms vs 30s)")
	fmt.Println("   ✅ Automatic recovery testing")
	fmt.Println("   ✅ Protects both application and infrastructure")
	fmt.Println("   ✅ Essential for production resilience")
}

func formatCircuitState(open bool) string {
	if open {
		return "🔴 OPEN"
	}
	return "🟢 CLOSED"
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
