// Package main demonstrates HTTP middleware for automatic context injection.
//
// The middleware automatically extracts user context from HTTP requests and
// makes it available for flag evaluation, eliminating boilerplate code.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🌐 Vexilla HTTP Middleware Example")
	fmt.Println(repeat("=", 80))
	fmt.Println()
	fmt.Println("This example demonstrates automatic context injection via HTTP middleware.")
	fmt.Println()

	// Create Vexilla client
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

	fmt.Println("✅ Vexilla client started successfully!")
	fmt.Println()

	// Create HTTP router
	mux := http.NewServeMux()

	// Example 1: Simple feature flag check
	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		// Extract user context from headers
		userID := r.Header.Get("X-User-ID")
		tier := r.Header.Get("X-User-Tier")

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("tier", tier)

		// Check if user has access to premium dashboard
		hasPremiumDashboard := client.Bool(r.Context(), "premium_dashboard", evalCtx)

		response := map[string]interface{}{
			"user_id":          userID,
			"tier":             tier,
			"premium_features": hasPremiumDashboard,
		}

		if hasPremiumDashboard {
			response["dashboard_url"] = "/dashboard/premium"
		} else {
			response["dashboard_url"] = "/dashboard/standard"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example 2: A/B testing
	mux.HandleFunc("/api/pricing", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		country := r.Header.Get("X-Country")

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("country", country)

		// Get pricing layout variant
		result, err := client.Evaluate(r.Context(), "pricing_layout", evalCtx)
		if err != nil {
			http.Error(w, "Failed to evaluate feature flag", http.StatusInternalServerError)
			return
		}

		layout := result.GetString("layout", "standard")

		// Return different pricing based on variant
		prices := map[string]interface{}{
			"standard": map[string]string{
				"basic":      "$9/month",
				"pro":        "$29/month",
				"enterprise": "$99/month",
			},
			"discount": map[string]string{
				"basic":      "$7/month",
				"pro":        "$24/month",
				"enterprise": "$79/month",
			},
			"premium": map[string]string{
				"basic":      "$12/month",
				"pro":        "$39/month",
				"enterprise": "$149/month",
			},
		}

		response := map[string]interface{}{
			"user_id": userID,
			"country": country,
			"layout":  layout,
			"prices":  prices[layout],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example 3: Feature gating
	mux.HandleFunc("/api/beta-features", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		tier := r.Header.Get("X-User-Tier")
		signupDate := r.Header.Get("X-Signup-Date")

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("tier", tier).
			WithAttribute("signup_date", signupDate)

		// Check beta access
		hasBetaAccess := client.Bool(r.Context(), "beta_access", evalCtx)

		if !hasBetaAccess {
			http.Error(w, "Beta access not available", http.StatusForbidden)
			return
		}

		// Return beta features
		features := []string{
			"advanced_analytics",
			"custom_themes",
			"api_access",
			"priority_support",
		}

		response := map[string]interface{}{
			"user_id":       userID,
			"beta_access":   true,
			"beta_features": features,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example 4: Regional features
	mux.HandleFunc("/api/features", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		country := r.Header.Get("X-Country")

		evalCtx := vexilla.NewContext(userID).
			WithAttribute("country", country)

		// Check multiple feature flags
		features := map[string]bool{
			"dark_mode":         client.Bool(r.Context(), "dark_mode", evalCtx),
			"new_ui":            client.Bool(r.Context(), "new_ui", evalCtx),
			"brazil_launch":     client.Bool(r.Context(), "brazil_launch", evalCtx),
			"payment_pix":       client.Bool(r.Context(), "payment_pix", evalCtx),
			"premium_features":  client.Bool(r.Context(), "premium_features", evalCtx),
		}

		response := map[string]interface{}{
			"user_id":  userID,
			"country":  country,
			"features": features,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Example 5: Middleware wrapper for automatic logging
	loggingMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			userID := r.Header.Get("X-User-ID")
			fmt.Printf("[%s] %s %s (user: %s)\n",
				time.Now().Format("15:04:05"),
				r.Method,
				r.URL.Path,
				userID,
			)

			next.ServeHTTP(w, r)

			duration := time.Since(start)
			fmt.Printf("  └─> Completed in %v\n", duration)
		}
	}

	// Wrap handlers with logging
	mux.HandleFunc("/api/logged-dashboard",
		loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("X-User-ID")
			evalCtx := vexilla.NewContext(userID)

			enabled := client.Bool(r.Context(), "new_feature", evalCtx)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"user_id":     userID,
				"new_feature": enabled,
			})
		}),
	)

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		metrics := client.Metrics()

		health := map[string]interface{}{
			"status":        "healthy",
			"circuit_open":  metrics.CircuitOpen,
			"last_refresh":  metrics.LastRefresh,
			"keys_cached":   metrics.Storage.KeysAdded,
			"hit_ratio":     metrics.Storage.HitRatio,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})

	// Start HTTP server
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	fmt.Println(repeat("-", 80))
	fmt.Println("🚀 HTTP Server Started")
	fmt.Println(repeat("-", 80))
	fmt.Printf("\nServer listening on http://localhost%s\n\n", addr)
	fmt.Println("Available endpoints:")
	fmt.Printf("  GET  http://localhost%s/api/dashboard\n", addr)
	fmt.Printf("  GET  http://localhost%s/api/pricing\n", addr)
	fmt.Printf("  GET  http://localhost%s/api/beta-features\n", addr)
	fmt.Printf("  GET  http://localhost%s/api/features\n", addr)
	fmt.Printf("  GET  http://localhost%s/health\n", addr)
	fmt.Println()
	fmt.Println("Example curl commands:")
	fmt.Println()
	fmt.Printf("  curl -H 'X-User-ID: user-123' -H 'X-User-Tier: premium' \\\n")
	fmt.Printf("       http://localhost%s/api/dashboard\n", addr)
	fmt.Println()
	fmt.Printf("  curl -H 'X-User-ID: user-456' -H 'X-Country: BR' \\\n")
	fmt.Printf("       http://localhost%s/api/pricing\n", addr)
	fmt.Println()
	fmt.Printf("  curl -H 'X-User-ID: user-789' -H 'X-User-Tier: premium' \\\n")
	fmt.Printf("       -H 'X-Signup-Date: 2025-01-01' \\\n")
	fmt.Printf("       http://localhost%s/api/beta-features\n", addr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println(repeat("-", 80))
	fmt.Println()

	// Start server
	log.Fatal(http.ListenAndServe(addr, mux))
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
