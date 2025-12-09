package vexilla

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// WebhookPayload represents the webhook notification from Flagr
type WebhookPayload struct {
	Event     string   `json:"event"`
	FlagKeys  []string `json:"flag_keys"`
	Timestamp string   `json:"timestamp"`
}

// startWebhookServer starts the HTTP server for receiving webhooks
func (c *Cache) startWebhookServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc(c.config.WebhookPath, c.handleWebhook)

	c.webhookServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.config.WebhookPort),
		Handler: mux,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.webhookServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("webhook server error: %v\n", err)
		}
	}()

	return nil
}

// handleWebhook processes incoming webhook notifications
func (c *Cache) handleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx, span := c.tracer.Start(r.Context(), "webhook.handle")
	defer span.End()

	if c.config.WebhookSecret != "" {
		secret := r.Header.Get("X-Webhook-Secret")
		if secret != c.config.WebhookSecret {
			span.AddEvent("unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	c.webhookEvents.Add(ctx, 1, metric.WithAttributes(
		attribute.String("event.type", payload.Event),
	))

	span.SetAttributes(
		attribute.String("event.type", payload.Event),
		attribute.StringSlice("flag.keys", payload.FlagKeys),
	)

	switch payload.Event {
	case "flag.updated":
		for _, key := range payload.FlagKeys {
			c.cache.Del(key)
		}
		go c.refreshFlags(context.Background())

	case "flag.deleted":
		for _, key := range payload.FlagKeys {
			c.cache.Del(key)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
