package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sign(secret string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func TestWebhook_HandleWebhook_Update(t *testing.T) {
	mock := &mockCache{}
	secret := "abc123"
	webhook := NewWebhookServer(mock, 0, secret)

	body := []byte(`{"event":"flag.updated","flag_keys":["a","b"]}`)
	sig := sign(secret, body)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
	req.Header.Set("X-Webhook-Signature", sig)

	w := httptest.NewRecorder()
	webhook.handleWebhook(w, req)

	assert.True(t, mock.InvalidateCalled)
	assert.True(t, mock.RefreshCalled)
}

func TestWebhook_InvalidSignature(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "secret")

	body := []byte(`{"event":"flag.updated","flag_keys":["x"]}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
	req.Header.Set("X-Webhook-Signature", "invalid")

	w := httptest.NewRecorder()
	webhook.handleWebhook(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, mock.InvalidateCalled)
}

func TestWebhook_Delete(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "")

	body := []byte(`{"event":"flag.deleted","flag_keys":["x"]}`)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	webhook.handleWebhook(w, req)

	assert.True(t, mock.InvalidateCalled)
	assert.False(t, mock.RefreshCalled)
}

func TestWebhook_InvalidJSON(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "")

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte("{invalid_json")))
	w := httptest.NewRecorder()

	webhook.handleWebhook(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhook_RequestBodyTooLarge(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "")

	// Create a valid JSON payload larger than MaxRequestBodySize (1MB)
	// Build a large flag_keys array
	flagKeys := make([]string, 0)
	currentSize := len(`{"event":"flag.updated","flag_keys":[]}`)

	// Add flag keys until we exceed the limit
	for currentSize < MaxRequestBodySize+1000 {
		key := "very_long_flag_key_name_with_lots_of_characters_0123456789"
		flagKeys = append(flagKeys, key)
		currentSize += len(key) + 3 // quotes and comma
	}

	payload := WebhookPayload{
		Event:    "flag.updated",
		FlagKeys: flagKeys,
	}
	largeJSON, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(largeJSON))
	w := httptest.NewRecorder()

	webhook.handleWebhook(w, req)

	// Should return 413 Request Entity Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.False(t, mock.InvalidateCalled)
	assert.False(t, mock.RefreshCalled)
}

func TestWebhook_RequestBodyAtLimit(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "")

	// Create a valid JSON payload close to the size limit
	// Build a large but valid JSON with many flag keys
	flagKeys := make([]string, 0)
	jsonSize := len(`{"event":"flag.updated","flag_keys":[]}`)

	for jsonSize < MaxRequestBodySize-1000 { // Leave room for JSON formatting
		key := "flag_key_with_long_name_000000"
		flagKeys = append(flagKeys, key)
		jsonSize += len(key) + 3 // quotes and comma
	}

	payload := WebhookPayload{
		Event:    "flag.updated",
		FlagKeys: flagKeys,
	}

	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/webhook", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	webhook.handleWebhook(w, req)

	// Should succeed even with a large (but under limit) payload
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mock.InvalidateCalled)
	assert.True(t, mock.RefreshCalled)
}

func TestWebhook_StartShutdown(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 18888, "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- webhook.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:18888/webhook")
	require.NoError(t, err)
	resp.Body.Close()

	// Should get method not allowed (webhook expects POST)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	err = webhook.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Wait for Start to return
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after shutdown")
	}
}

func TestWebhook_Shutdown_NotStarted(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 0, "")

	// Shutdown without starting should not error
	err := webhook.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestWebhook_ContextCancellation(t *testing.T) {
	mock := &mockCache{}
	webhook := NewWebhookServer(mock, 18889, "")

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- webhook.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for Start to return
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}
