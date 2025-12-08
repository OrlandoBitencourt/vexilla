package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

func main() {
	// Create configuration
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000" // Flagr server
	config.RefreshInterval = 2 * time.Minute
	config.PersistenceEnabled = true
	config.PersistencePath = "/tmp/vexilla-demo"

	// Enable webhook for instant updates
	config.WebhookEnabled = true
	config.WebhookPort = 8081
	config.WebhookSecret = "my-webhook-secret"

	// Enable admin API
	config.AdminAPIEnabled = true
	config.AdminAPIPort = 8082

	log.Println("🏴 Starting Vexilla...")
	log.Printf("Flagr endpoint: %s", config.FlagrEndpoint)
	log.Printf("Webhook: http://localhost:%d%s", config.WebhookPort, config.WebhookPath)
	log.Printf("Admin API: http://localhost:%d%s", config.AdminAPIPort, config.AdminAPIPath)

	// Create Vexilla cache (implementation in next artifact)
	// cache, err := vexilla.New(config)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Start background services
	// if err := cache.Start(); err != nil {
	//     log.Fatal(err)
	// }
	// defer cache.Stop()

	// Example evaluations
	ctx := context.Background()
	defer ctx.Done()

	// Example 1: Static flag (evaluated locally)
	// result := cache.EvaluateBool(ctx, "brazil_launch", vexilla.EvaluationContext{
	//     EntityID: "user123",
	//     Context: map[string]interface{}{
	//         "country":  "BR",
	//         "document": "1234567808",
	//     },
	// })
	// log.Printf("Brazil launch enabled: %v (local evaluation)", result)

	// Example 2: Dynamic flag (requires Flagr)
	// result, err := cache.Evaluate(ctx, "ab_test", vexilla.EvaluationContext{
	//     EntityID: "user456",
	//     Context: map[string]interface{}{
	//         "country": "US",
	//     },
	// })
	// if err == nil {
	//     log.Printf("A/B test variant: %s (Flagr evaluation)", result.VariantKey)
	// }

	// Start HTTP server with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/feature", featureHandler)

	// Wrap with Vexilla middleware (uncomment when cache is ready)
	// handler := cache.Middleware(mux)

	log.Println("🚀 Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Example: Get flag from context
	// if vexilla.GetFlagBoolFromContext(r.Context(), "new_ui") {
	//     fmt.Fprintf(w, "<h1>New UI!</h1>")
	// } else {
	//     fmt.Fprintf(w, "<h1>Classic UI</h1>")
	// }
	fmt.Fprintf(w, "<h1>Vexilla Demo</h1>")
}

func featureHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer ctx.Done()

	// Multiple flag evaluations
	features := map[string]bool{
		"premium": false, // vexilla.GetFlagBoolFromContext(ctx, "premium")
		"beta":    false, // vexilla.GetFlagBoolFromContext(ctx, "beta")
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"features": %v}`, features)
}
