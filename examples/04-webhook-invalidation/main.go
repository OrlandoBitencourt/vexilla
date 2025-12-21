// Package main demonstrates webhook-based cache invalidation for real-time flag updates.
//
// This example shows how to receive real-time updates from Flagr when flags are
// modified, enabling sub-second cache invalidation without waiting for background refresh.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🔔 Vexilla Webhook Invalidation Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates real-time cache invalidation via webhooks.")
	fmt.Println("Flagr will notify Vexilla when flags are updated, enabling instant updates.")
	fmt.Println()

	// Configuration
	webhookPort := 18001
	webhookSecret := "my-webhook-secret-key" // Shared secret with Flagr

	fmt.Printf("🔧 Configuration:\n")
	fmt.Printf("   Webhook Port: %d\n", webhookPort)
	fmt.Printf("   Webhook Secret: %s\n", webhookSecret)
	fmt.Printf("   Flagr Endpoint: http://localhost:18000\n")
	fmt.Println()

	// Create client with webhook invalidation enabled
	fmt.Println("📦 Creating Vexilla client with webhook support...")
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),
		vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
			Port:   webhookPort,
			Secret: webhookSecret,
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
	fmt.Printf("🔔 Webhook server listening on http://localhost:%d/webhook\n", webhookPort)
	fmt.Println()

	// Setup instructions
	fmt.Println(repeat("-", 80))
	fmt.Println("🔧 SETUP INSTRUCTIONS")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("To enable webhook notifications in Flagr, add this configuration:")
	fmt.Println()
	fmt.Println("1. Edit your Flagr configuration file:")
	fmt.Println()
	fmt.Println("   webhooks:")
	fmt.Println("     - url: http://localhost:18001/webhook")
	fmt.Printf("       secret: %s\n", webhookSecret)
	fmt.Println("       events:")
	fmt.Println("         - flag.updated")
	fmt.Println("         - flag.deleted")
	fmt.Println()
	fmt.Println("2. Restart Flagr to apply changes")
	fmt.Println()
	fmt.Println("3. Now when you modify a flag in Flagr UI, the webhook will be triggered!")
	fmt.Println()
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Initial flag evaluation
	fmt.Println("📊 Initial Flag State")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	testFlags := []string{
		"new_feature",
		"premium_features",
		"dark_mode",
	}

	evalCtx := vexilla.NewContext("webhook-test-user")

	for _, flagKey := range testFlags {
		enabled := client.Bool(ctx, flagKey, evalCtx)
		status := formatStatus(enabled)
		fmt.Printf("%-20s: %s\n", flagKey, status)
	}
	fmt.Println()

	// Monitor flag changes
	fmt.Println("🔄 Monitoring for Flag Changes")
	fmt.Println(repeat("-", 80))
	fmt.Println()
	fmt.Println("Now modify a flag in Flagr UI and watch for real-time updates!")
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println()

	// Create ticker to periodically check flag values
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Track previous values to detect changes
	previousValues := make(map[string]bool)
	for _, flagKey := range testFlags {
		previousValues[flagKey] = client.Bool(ctx, flagKey, evalCtx)
	}

	changeCount := 0

	// Monitoring loop
	for {
		select {
		case <-ticker.C:
			// Check each flag for changes
			for _, flagKey := range testFlags {
				currentValue := client.Bool(ctx, flagKey, evalCtx)
				previousValue := previousValues[flagKey]

				if currentValue != previousValue {
					changeCount++
					timestamp := time.Now().Format("15:04:05")

					fmt.Printf("[%s] 🔔 FLAG CHANGED: %s\n", timestamp, flagKey)
					fmt.Printf("           Previous: %s\n", formatStatus(previousValue))
					fmt.Printf("           Current:  %s\n", formatStatus(currentValue))
					fmt.Printf("           Change #%d detected via webhook!\n", changeCount)
					fmt.Println()

					previousValues[flagKey] = currentValue
				}
			}

		case <-sigChan:
			fmt.Println()
			fmt.Println("👋 Shutting down gracefully...")
			fmt.Println()

			// Show final statistics
			fmt.Println("📊 Session Summary")
			fmt.Println(repeat("-", 80))
			fmt.Printf("Total flag changes detected: %d\n", changeCount)
			fmt.Println()

			metrics := client.Metrics()
			fmt.Printf("Cache statistics:\n")
			fmt.Printf("  Keys cached: %d\n", metrics.Storage.KeysAdded)
			fmt.Printf("  Hit ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
			fmt.Printf("  Last refresh: %v ago\n", time.Since(metrics.LastRefresh).Round(time.Second))
			fmt.Println()

			return
		}
	}
}

func formatStatus(enabled bool) string {
	if enabled {
		return "✅ ENABLED"
	}
	return "❌ DISABLED"
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
