package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

func main() {
	// Create and start Vexilla
	config := vexilla.DefaultConfig()
	config.FlagrEndpoint = "http://localhost:18000"
	config.AdminAPIEnabled = true
	config.AdminAPIPort = 8082

	log.Println("🏴 Creating Vexilla client...")
	client, err := vexilla.New(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("🚀 Starting Vexilla...")
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Setup HTTP server with Vexilla middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/api/features", featuresHandler)

	// Wrap with Vexilla middleware
	handler := client.Middleware(mux)

	log.Println("🚀 Server starting on :8080")
	log.Println("📊 Admin API on :8082/admin")
	log.Println("")
	log.Println("Try:")
	log.Println("  curl http://localhost:8080")
	log.Println("  curl http://localhost:8080/api/features")
	log.Println("  curl http://localhost:8082/admin/stats")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get flag from context
	if vexilla.GetFlagBoolFromContext(r.Context(), "new_ui") {
		fmt.Fprintf(w, "<h1>🎨 New UI Design!</h1>")
		fmt.Fprintf(w, "<p>You're seeing the new interface</p>")
	} else {
		fmt.Fprintf(w, "<h1>Classic UI</h1>")
		fmt.Fprintf(w, "<p>You're seeing the classic interface</p>")
	}
}

func featuresHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Evaluate multiple flags
	features := map[string]interface{}{
		"premium": vexilla.GetFlagBoolFromContext(ctx, "premium_features"),
		"beta":    vexilla.GetFlagBoolFromContext(ctx, "beta_features"),
		"new_ui":  vexilla.GetFlagBoolFromContext(ctx, "new_ui"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"features": features,
		"user_id":  r.Header.Get("X-User-ID"),
	})
}
