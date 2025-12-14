package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WebhookServer handles webhooks from Flagr
type WebhookServer struct {
	cache  CacheInterface
	port   int
	secret string
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

// Start starts the webhook HTTP server
func (w *WebhookServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", w.handleWebhook)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (w *WebhookServer) handleWebhook(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
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
