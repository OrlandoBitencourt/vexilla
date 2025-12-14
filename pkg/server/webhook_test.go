package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
