// Package main demonstrates OpenTelemetry integration for observability.
//
// This example shows how to enable OpenTelemetry tracing and metrics
// for comprehensive observability of Vexilla operations.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("📊 Vexilla Telemetry Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates OpenTelemetry integration for observability.")
	fmt.Println()

	// Example 1: Basic metrics
	fmt.Println("Example 1: Built-in Metrics")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("Creating Vexilla client...")
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

	// Perform various operations to generate metrics
	fmt.Println("Performing operations to generate metrics...")
	fmt.Println()

	evalCtx := vexilla.NewContext("telemetry-test-user").
		WithAttribute("tier", "premium").
		WithAttribute("country", "BR")

	testFlags := []string{
		"new_feature",
		"premium_features",
		"dark_mode",
		"beta_access",
		"brazil_launch",
	}

	// Perform multiple evaluations
	fmt.Println("Evaluating flags (10 iterations)...")
	for i := 0; i < 10; i++ {
		for _, flagKey := range testFlags {
			_ = client.Bool(ctx, flagKey, evalCtx)
		}
	}
	fmt.Printf("✅ Completed %d evaluations\n", 10*len(testFlags))
	fmt.Println()

	// Example 2: Cache metrics
	fmt.Println("Example 2: Cache Performance Metrics")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	metrics := client.Metrics()

	fmt.Println("📦 Storage Metrics:")
	fmt.Printf("  Keys Added:     %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  Keys Evicted:   %d\n", metrics.Storage.KeysEvicted)
	fmt.Printf("  Hit Ratio:      %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Println()

	// Example 3: Health metrics
	fmt.Println("Example 3: Health Status Metrics")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("🏥 Health Metrics:")
	fmt.Printf("  Last Refresh:        %v ago\n", time.Since(metrics.LastRefresh).Round(time.Second))
	fmt.Printf("  Circuit Breaker:     %s\n", formatCircuitState(metrics.CircuitOpen))
	fmt.Printf("  Consecutive Fails:   %d\n", metrics.ConsecutiveFails)
	fmt.Println()

	// Example 4: Performance analysis
	fmt.Println("Example 4: Performance Analysis")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Benchmark local evaluations
	fmt.Println("Benchmarking local evaluations...")
	localStart := time.Now()
	localIterations := 10000

	for i := 0; i < localIterations; i++ {
		userID := fmt.Sprintf("user-%d", i)
		ctx := vexilla.NewContext(userID)
		_ = client.Bool(context.Background(), "new_feature", ctx)
	}

	localDuration := time.Since(localStart)
	localAvg := localDuration / time.Duration(localIterations)

	fmt.Printf("Local Evaluations:\n")
	fmt.Printf("  Total:          %d evaluations\n", localIterations)
	fmt.Printf("  Duration:       %v\n", localDuration)
	fmt.Printf("  Average:        %v per evaluation\n", localAvg)
	fmt.Printf("  Throughput:     ~%d eval/sec\n", int(float64(localIterations)/localDuration.Seconds()))
	fmt.Println()

	// Example 5: Cache efficiency
	fmt.Println("Example 5: Cache Efficiency Analysis")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	updatedMetrics := client.Metrics()

	hitRatio := updatedMetrics.Storage.HitRatio
	hitRate := hitRatio * 100

	fmt.Println("💾 Cache Efficiency:")
	fmt.Printf("  Hit Ratio:      %.2f%%\n", hitRate)

	if hitRate >= 95 {
		fmt.Println("  Status:         ✅ EXCELLENT")
	} else if hitRate >= 85 {
		fmt.Println("  Status:         ✅ GOOD")
	} else if hitRate >= 70 {
		fmt.Println("  Status:         ⚠️  FAIR")
	} else {
		fmt.Println("  Status:         ❌ POOR")
	}

	fmt.Println()

	// Example 6: Metrics over time
	fmt.Println("Example 6: Metrics Monitoring Over Time")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("Monitoring metrics for 20 seconds...")
	fmt.Println()

	fmt.Printf("%-10s | %-12s | %-10s | %-15s | %-10s\n",
		"Time", "Keys Cached", "Hit Ratio", "Circuit State", "Fails")
	fmt.Println(repeat("-", 80))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	iterations := 4
	for i := 0; i < iterations; i++ {
		// Perform some evaluations
		for j := 0; j < 5; j++ {
			userID := fmt.Sprintf("monitor-user-%d", i*5+j)
			ctx := vexilla.NewContext(userID)
			_ = client.Bool(context.Background(), testFlags[i%len(testFlags)], ctx)
		}

		<-ticker.C

		m := client.Metrics()
		timestamp := time.Now().Format("15:04:05")
		circuitState := formatCircuitStateShort(m.CircuitOpen)

		fmt.Printf("%-10s | %-12d | %9.2f%% | %-15s | %-10d\n",
			timestamp,
			m.Storage.KeysAdded,
			m.Storage.HitRatio*100,
			circuitState,
			m.ConsecutiveFails,
		)
	}

	fmt.Println()

	// Example 7: OpenTelemetry integration (conceptual)
	fmt.Println("Example 7: OpenTelemetry Integration (Conceptual)")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("📊 Vexilla exports the following OpenTelemetry metrics:")
	fmt.Println()

	fmt.Println("Counters:")
	fmt.Println("  • vexilla.cache.hits          - Cache hit counter")
	fmt.Println("  • vexilla.cache.misses        - Cache miss counter")
	fmt.Println("  • vexilla.evaluations         - Total evaluations (with strategy label)")
	fmt.Println("  • vexilla.refresh.success     - Successful refresh counter")
	fmt.Println("  • vexilla.refresh.failure     - Failed refresh counter")
	fmt.Println()

	fmt.Println("Histograms:")
	fmt.Println("  • vexilla.refresh.duration    - Refresh latency distribution")
	fmt.Println("  • vexilla.evaluation.duration - Evaluation latency distribution")
	fmt.Println()

	fmt.Println("Gauges:")
	fmt.Println("  • vexilla.circuit.state       - Circuit breaker state (0=closed, 1=open)")
	fmt.Println("  • vexilla.cache.size          - Number of cached flags")
	fmt.Println("  • vexilla.cache.hit_ratio     - Cache hit ratio percentage")
	fmt.Println()

	// Example 8: Monitoring recommendations
	fmt.Println("Example 8: Monitoring Dashboard Recommendations")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("📈 Recommended Dashboards:")
	fmt.Println()

	fmt.Println("1. Cache Performance:")
	fmt.Println("   • Hit ratio over time")
	fmt.Println("   • Cache size trend")
	fmt.Println("   • Eviction rate")
	fmt.Println()

	fmt.Println("2. Evaluation Metrics:")
	fmt.Println("   • Total evaluations per second")
	fmt.Println("   • Local vs remote distribution")
	fmt.Println("   • Evaluation latency percentiles (p50, p95, p99)")
	fmt.Println()

	fmt.Println("3. Reliability:")
	fmt.Println("   • Circuit breaker state")
	fmt.Println("   • Refresh success rate")
	fmt.Println("   • Consecutive failures")
	fmt.Println("   • Time since last successful refresh")
	fmt.Println()

	fmt.Println("4. Resource Usage:")
	fmt.Println("   • Memory usage (cache size × avg flag size)")
	fmt.Println("   • Network requests to Flagr")
	fmt.Println("   • Disk I/O (if disk persistence enabled)")
	fmt.Println()

	// Example 9: Alerting rules
	fmt.Println("Example 9: Recommended Alerting Rules")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	fmt.Println("🚨 Critical Alerts:")
	fmt.Println("   • Circuit breaker open for > 5 minutes")
	fmt.Println("   • Cache hit ratio < 50%")
	fmt.Println("   • No successful refresh in 15 minutes")
	fmt.Println("   • Consecutive failures >= 5")
	fmt.Println()

	fmt.Println("⚠️  Warning Alerts:")
	fmt.Println("   • Circuit breaker opened (transient)")
	fmt.Println("   • Cache hit ratio < 85%")
	fmt.Println("   • Refresh duration > 5 seconds")
	fmt.Println("   • Consecutive failures >= 2")
	fmt.Println()

	fmt.Println(repeat("=", 80))
	fmt.Println("✅ Telemetry example completed!")
	fmt.Println()
	fmt.Println("🔑 Key Metrics to Monitor:")
	fmt.Println("   📊 Cache hit ratio (target: >95%)")
	fmt.Println("   ⚡ Evaluation latency (target: <1ms for local)")
	fmt.Println("   🔄 Refresh success rate (target: 100%)")
	fmt.Println("   🔌 Circuit breaker state (target: closed)")
	fmt.Println()
	fmt.Println("💡 Integration Tips:")
	fmt.Println("   • Export metrics to Prometheus")
	fmt.Println("   • Visualize in Grafana")
	fmt.Println("   • Set up alerts for anomalies")
	fmt.Println("   • Track metrics in production")
}

func formatCircuitState(open bool) string {
	if open {
		return "🔴 OPEN (degraded)"
	}
	return "🟢 CLOSED (healthy)"
}

func formatCircuitStateShort(open bool) string {
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
