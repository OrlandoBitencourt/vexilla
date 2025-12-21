// Package main demonstrates the Admin Server for cache management and monitoring.
//
// The Admin Server provides a REST API for operational tasks like health checks,
// metrics, cache invalidation, and manual refresh.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("⚙️  Vexilla Admin Server Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates the Admin Server for cache management.")
	fmt.Println()

	// Configuration
	adminPort := 19000

	fmt.Printf("🔧 Configuration:\n")
	fmt.Printf("   Admin Port: %d\n", adminPort)
	fmt.Printf("   Flagr Endpoint: http://localhost:18000\n")
	fmt.Println()

	// Create client with admin server enabled
	fmt.Println("📦 Creating Vexilla client with Admin Server...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithAdminServer(vexilla.AdminConfig{
			Port: adminPort,
		}),
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
	fmt.Printf("🌐 Admin server running on http://localhost:%d\n", adminPort)
	fmt.Println()

	// Display available endpoints
	fmt.Println(repeat("-", 80))
	fmt.Println("📚 AVAILABLE ENDPOINTS")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("Health Check:")
	fmt.Printf("  GET  http://localhost:%d/health\n", adminPort)
	fmt.Println()
	fmt.Println("Cache Statistics:")
	fmt.Printf("  GET  http://localhost:%d/admin/stats\n", adminPort)
	fmt.Println()
	fmt.Println("Cache Management:")
	fmt.Printf("  POST http://localhost:%d/admin/invalidate        (invalidate specific flag)\n", adminPort)
	fmt.Printf("  POST http://localhost:%d/admin/invalidate-all    (clear entire cache)\n", adminPort)
	fmt.Printf("  POST http://localhost:%d/admin/refresh           (force refresh)\n", adminPort)
	fmt.Println()
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Demonstrate each endpoint
	baseURL := fmt.Sprintf("http://localhost:%d", adminPort)

	// Wait a bit for server to fully start
	time.Sleep(500 * time.Millisecond)

	// Example 1: Health Check
	fmt.Println("Example 1: Health Check")
	fmt.Println(repeat("-", 80))
	healthCheck(baseURL)
	fmt.Println()

	// Example 2: Get Statistics
	fmt.Println("Example 2: Cache Statistics")
	fmt.Println(repeat("-", 80))
	getStats(baseURL)
	fmt.Println()

	// Perform some evaluations to populate cache
	fmt.Println("Example 3: Populate Cache with Evaluations")
	fmt.Println(repeat("-", 80))
	fmt.Println("Performing test evaluations...")

	testFlags := []string{
		"new_feature",
		"premium_features",
		"dark_mode",
		"beta_access",
	}

	evalCtx := vexilla.NewContext("admin-test-user").
		WithAttribute("tier", "premium")

	for _, flagKey := range testFlags {
		enabled := client.Bool(ctx, flagKey, evalCtx)
		status := "❌"
		if enabled {
			status = "✅"
		}
		fmt.Printf("  %s %s\n", status, flagKey)
	}
	fmt.Println()

	// Example 4: Get updated statistics
	fmt.Println("Example 4: Updated Cache Statistics")
	fmt.Println(repeat("-", 80))
	getStats(baseURL)
	fmt.Println()

	// Example 5: Invalidate specific flag
	fmt.Println("Example 5: Invalidate Specific Flag")
	fmt.Println(repeat("-", 80))
	invalidateFlag(baseURL, "new_feature")
	fmt.Println()

	// Example 6: Force refresh
	fmt.Println("Example 6: Force Refresh All Flags")
	fmt.Println(repeat("-", 80))
	forceRefresh(baseURL)
	fmt.Println()

	// Example 7: Invalidate all
	fmt.Println("Example 7: Invalidate All Flags")
	fmt.Println(repeat("-", 80))
	fmt.Println("⚠️  Warning: This will clear the entire cache!")
	fmt.Println("Clearing cache in 2 seconds...")
	time.Sleep(2 * time.Second)
	invalidateAll(baseURL)
	fmt.Println()

	// Example 8: Stats after clearing
	fmt.Println("Example 8: Statistics After Cache Clear")
	fmt.Println(repeat("-", 80))
	getStats(baseURL)
	fmt.Println()

	// Interactive mode
	fmt.Println(repeat("=", 80))
	fmt.Println("🎮 INTERACTIVE MODE")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("The Admin Server is now running. You can interact with it using:")
	fmt.Println()
	fmt.Println("curl commands:")
	fmt.Printf("  curl http://localhost:%d/health\n", adminPort)
	fmt.Printf("  curl http://localhost:%d/admin/stats\n", adminPort)
	fmt.Printf("  curl -X POST http://localhost:%d/admin/refresh\n", adminPort)
	fmt.Println()
	fmt.Println("Or use tools like:")
	fmt.Println("  • Postman")
	fmt.Println("  • HTTPie")
	fmt.Println("  • Browser (for GET endpoints)")
	fmt.Println()
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("👋 Shutting down gracefully...")
	fmt.Println()

	// Final statistics
	fmt.Println("📊 Final Session Summary")
	fmt.Println(repeat("-", 80))
	getStats(baseURL)

	fmt.Println()
	fmt.Println("✅ Example completed!")
}

func healthCheck(baseURL string) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response:\n%s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Service is healthy!")
	}
}

func getStats(baseURL string) {
	resp, err := http.Get(baseURL + "/admin/stats")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Pretty print JSON
	var stats map[string]interface{}
	if err := json.Unmarshal(body, &stats); err == nil {
		prettyJSON, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(prettyJSON))
	} else {
		fmt.Println(string(body))
	}
}

func invalidateFlag(baseURL, flagKey string) {
	payload := map[string]string{
		"flag_key": flagKey,
	}

	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(
		baseURL+"/admin/invalidate",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Invalidating flag: %s\n", flagKey)
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Flag '%s' invalidated successfully!\n", flagKey)
	}
}

func invalidateAll(baseURL string) {
	resp, err := http.Post(baseURL+"/admin/invalidate-all", "application/json", nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ All flags invalidated successfully!")
	}
}

func forceRefresh(baseURL string) {
	resp, err := http.Post(baseURL+"/admin/refresh", "application/json", nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ Cache refreshed successfully!")
	}
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
