package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AdminHandler handles admin operations
type AdminHandler interface {
	GetStats(ctx context.Context) (interface{}, error)
	InvalidateFlags(ctx context.Context, flagKeys []string) error
	InvalidateAll(ctx context.Context) error
	ForceRefresh(ctx context.Context) error
	HealthCheck(ctx context.Context) (interface{}, error)
}

// AdminServer serves admin API endpoints
type AdminServer struct {
	port    int
	path    string
	handler AdminHandler
	server  *http.Server
	tracer  trace.Tracer
	wg      sync.WaitGroup
}

// NewAdminServer creates a new admin server
func NewAdminServer(port int, path string, handler AdminHandler) *AdminServer {
	return &AdminServer{
		port:    port,
		path:    path,
		handler: handler,
		tracer:  otel.Tracer("vexilla.admin"),
	}
}

// Start starts the admin server
func (a *AdminServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	basePath := a.path
	mux.HandleFunc(basePath+"/stats", a.handleStats)
	mux.HandleFunc(basePath+"/invalidate", a.handleInvalidate)
	mux.HandleFunc(basePath+"/invalidate-all", a.handleInvalidateAll)
	mux.HandleFunc(basePath+"/refresh", a.handleRefresh)
	mux.HandleFunc("/health", a.handleHealth)

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: mux,
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("admin server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the admin server
func (a *AdminServer) Stop(ctx context.Context) error {
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	a.wg.Wait()
	return nil
}

func (a *AdminServer) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "admin.stats")
	defer span.End()

	stats, err := a.handler.GetStats(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (a *AdminServer) handleInvalidate(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "admin.invalidate")
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

	if err := a.handler.InvalidateFlags(ctx, req.FlagKeys); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "ok",
		"invalidated": req.FlagKeys,
	})
}

func (a *AdminServer) handleInvalidateAll(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "admin.invalidate_all")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := a.handler.InvalidateAll(ctx); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *AdminServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "admin.refresh")
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := a.handler.ForceRefresh(ctx); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "admin.health")
	defer span.End()

	health, err := a.handler.HealthCheck(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
