package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// MaxRequestBodySize is the maximum allowed size for request bodies (1MB)
	// This prevents DoS attacks from exhausting memory with large payloads
	MaxRequestBodySize = 1 << 20 // 1MB

	// DefaultReadTimeout is the maximum duration for reading the entire request
	DefaultReadTimeout = 10 * time.Second

	// DefaultWriteTimeout is the maximum duration before timing out writes
	DefaultWriteTimeout = 10 * time.Second

	// DefaultIdleTimeout is the maximum duration to wait for the next request
	DefaultIdleTimeout = 60 * time.Second
)

// WebhookServer handles webhooks from Flagr
type WebhookServer struct {
	cache  CacheInterface
	port   int
	secret string
	server *http.Server
	mu     sync.Mutex
}

// WebhookPayload represents the webhook payload from Flagr
type WebhookPayload struct {
	Event     string   `json:"event"`
	FlagKeys  []string `json:"flag_keys"`
	Timestamp string   `json:"timestamp"`
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(cache CacheInterface, port int, secret string) *WebhookServer {
	return &WebhookServer{
		cache:  cache,
		port:   port,
		secret: secret,
	}
}

// Start starts the webhook HTTP server with security timeouts.
// This method blocks until the server is stopped or an error occurs.
// Use the provided context to signal graceful shutdown.
func (w *WebhookServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", w.handleWebhook)

	w.mu.Lock()
	w.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", w.port),
		Handler:      mux,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}
	server := w.server
	w.mu.Unlock()

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

// Shutdown gracefully shuts down the webhook server.
// It waits for active connections to complete within the timeout.
func (w *WebhookServer) Shutdown(ctx context.Context) error {
	w.mu.Lock()
	server := w.server
	w.mu.Unlock()

	if server == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

func (w *WebhookServer) handleWebhook(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(rw, r.Body, MaxRequestBodySize)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Check if error is due to body being too large
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(rw, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(rw, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Verify signature if secret is configured
	if w.secret != "" {
		if !w.verifySignature(r, body) {
			http.Error(rw, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(rw, "Invalid JSON", http.StatusBadRequest)
		return
	}

	w.handleEvent(payload)

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{"status": "ok"})
}

func (w *WebhookServer) verifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(w.secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (w *WebhookServer) handleEvent(payload WebhookPayload) {
	switch payload.Event {
	case "flag.updated":
		for _, key := range payload.FlagKeys {
			w.cache.InvalidateFlag(key)
		}
		w.cache.RefreshFlags()

	case "flag.deleted":
		for _, key := range payload.FlagKeys {
			w.cache.InvalidateFlag(key)
		}
	}
}
