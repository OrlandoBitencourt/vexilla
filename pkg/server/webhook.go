package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// WebhookPayload represents incoming webhook data from Flagr
type WebhookPayload struct {
	Event     string   `json:"event"` // "flag.updated", "flag.deleted"
	FlagKeys  []string `json:"flag_keys"`
	Timestamp string   `json:"timestamp"`
}

// WebhookHandler handles flag change notifications
type WebhookHandler interface {
	OnFlagUpdated(ctx context.Context, flagKeys []string) error
	OnFlagDeleted(ctx context.Context, flagKeys []string) error
}

// WebhookServer serves webhook endpoint
type WebhookServer struct {
	port         int
	path         string
	secret       string
	handler      WebhookHandler
	server       *http.Server
	tracer       trace.Tracer
	eventCounter metric.Int64Counter
	wg           sync.WaitGroup
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer(port int, path string, secret string, handler WebhookHandler) (*WebhookServer, error) {
	tracer := otel.Tracer("vexilla.webhook")
	meter := otel.Meter("vexilla")

	eventCounter, err := meter.Int64Counter(
		"vexilla.webhook.events",
		metric.WithDescription("Webhook events received"),
	)
	if err != nil {
		return nil, err
	}

	return &WebhookServer{
		port:         port,
		path:         path,
		secret:       secret,
		handler:      handler,
		tracer:       tracer,
		eventCounter: eventCounter,
	}, nil
}

// Start starts the webhook server
func (w *WebhookServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc(w.path, w.handleWebhook)

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("webhook server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the webhook server
func (w *WebhookServer) Stop(ctx context.Context) error {
	if w.server != nil {
		if err := w.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	w.wg.Wait()
	return nil
}

// handleWebhook processes incoming webhooks
func (w *WebhookServer) handleWebhook(rw http.ResponseWriter, r *http.Request) {
	ctx, span := w.tracer.Start(r.Context(), "webhook.handle")
	defer span.End()

	// Verify secret if configured
	if w.secret != "" {
		secret := r.Header.Get("X-Webhook-Secret")
		if secret != w.secret {
			span.AddEvent("unauthorized")
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Parse payload
	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		span.RecordError(err)
		http.Error(rw, "Invalid payload", http.StatusBadRequest)
		return
	}

	w.eventCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("event.type", payload.Event),
	))

	span.SetAttributes(
		attribute.String("event.type", payload.Event),
		attribute.StringSlice("flag.keys", payload.FlagKeys),
	)

	// Handle event
	var err error
	switch payload.Event {
	case "flag.updated":
		err = w.handler.OnFlagUpdated(ctx, payload.FlagKeys)
	case "flag.deleted":
		err = w.handler.OnFlagDeleted(ctx, payload.FlagKeys)
	default:
		http.Error(rw, "Unknown event type", http.StatusBadRequest)
		return
	}

	if err != nil {
		span.RecordError(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(map[string]string{"status": "ok"})
}
