package vexilla

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// startAdminAPI starts the admin HTTP API
func (c *Cache) startAdminAPI() error {
	mux := http.NewServeMux()

	basePath := c.config.AdminAPIPath
	mux.HandleFunc(basePath+"/stats", c.handleAdminStats)
	mux.HandleFunc(basePath+"/invalidate", c.handleAdminInvalidate)
	mux.HandleFunc(basePath+"/invalidate-all", c.handleAdminInvalidateAll)
	mux.HandleFunc(basePath+"/refresh", c.handleAdminRefresh)
	mux.HandleFunc("/health", c.handleHealth)

	c.adminServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.config.AdminAPIPort),
		Handler: mux,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("admin API server error: %v\n", err)
		}
	}()

	return nil
}

func (c *Cache) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	stats := c.GetCacheStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (c *Cache) handleAdminInvalidate(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "admin.invalidate")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FlagKeys []string `json:"flag_keys"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.StringSlice("flag.keys", req.FlagKeys))

	for _, key := range req.FlagKeys {
		c.InvalidateFlag(ctx, key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"invalidated": req.FlagKeys,
	})
}

func (c *Cache) handleAdminInvalidateAll(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "admin.invalidate_all")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c.InvalidateAll(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (c *Cache) handleAdminRefresh(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "admin.refresh")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := c.refreshFlags(ctx); err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Refresh failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (c *Cache) handleHealth(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := map[string]interface{}{
		"status":       "healthy",
		"circuit_open": c.circuitOpen,
		"last_refresh": c.lastRefresh.Format(time.RFC3339),
	}

	if c.circuitOpen {
		health["status"] = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
