package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AdminServer provides admin HTTP endpoints
type AdminServer struct {
	cache CacheInterface
	port  int
}

// CacheInterface defines what the admin server needs from cache
type CacheInterface interface {
	GetMetrics() interface{}
	InvalidateFlag(flagKey string) error
	InvalidateAll() error
	RefreshFlags() error
}

// NewAdminServer creates a new admin server
func NewAdminServer(cache CacheInterface, port int) *AdminServer {
	return &AdminServer{
		cache: cache,
		port:  port,
	}
}

// Start starts the admin HTTP server
func (a *AdminServer) Start() error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", a.handleHealth)

	// Metrics
	mux.HandleFunc("/admin/stats", a.handleStats)

	// Cache management
	mux.HandleFunc("/admin/invalidate", a.handleInvalidate)
	mux.HandleFunc("/admin/invalidate-all", a.handleInvalidateAll)
	mux.HandleFunc("/admin/refresh", a.handleRefresh)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (a *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (a *AdminServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := a.cache.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (a *AdminServer) handleInvalidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FlagKey string `json:"flag_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := a.cache.InvalidateFlag(req.FlagKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"flag":   req.FlagKey,
	})
}

func (a *AdminServer) handleInvalidateAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := a.cache.InvalidateAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *AdminServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := a.cache.RefreshFlags(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
