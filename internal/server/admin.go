package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AdminServer provides admin HTTP endpoints
type AdminServer struct {
	cache  CacheInterface
	port   int
	server *http.Server
	mu     sync.Mutex
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

// Start starts the admin HTTP server with security timeouts.
// This method blocks until the server is stopped or an error occurs.
// Use the provided context to signal graceful shutdown.
func (a *AdminServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", a.handleHealth)

	// Metrics
	mux.HandleFunc("/admin/stats", a.handleStats)

	// Cache management
	mux.HandleFunc("/admin/invalidate", a.handleInvalidate)
	mux.HandleFunc("/admin/invalidate-all", a.handleInvalidateAll)
	mux.HandleFunc("/admin/refresh", a.handleRefresh)

	a.mu.Lock()
	a.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.port),
		Handler:      mux,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	server := a.server
	a.mu.Unlock()

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		} else {
			// Server stopped gracefully, signal completion
			errChan <- nil
		}
	}()

	// Wait for context cancellation or server completion
	select {
	case <-ctx.Done():
		// Context canceled, shutdown server if not already done
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the admin server.
// It waits for active connections to complete within the timeout.
func (a *AdminServer) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	server := a.server
	a.mu.Unlock()

	if server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
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

	// Limit request body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

	var req struct {
		FlagKey string `json:"flag_key"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		// Check if error is due to body being too large
		// MaxBytesReader returns MaxBytesError which can be detected using errors.As
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
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
