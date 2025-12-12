package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ----------------------
// Mock Cache
// ----------------------

type mockCache struct {
	InvalidateCalled    bool
	InvalidateAllCalled bool
	RefreshCalled       bool
	LastInvalidatedKey  string
	Metrics             interface{}
}

func (m *mockCache) GetMetrics() interface{} {
	return m.Metrics
}

func (m *mockCache) InvalidateFlag(flagKey string) error {
	m.InvalidateCalled = true
	m.LastInvalidatedKey = flagKey
	return nil
}

func (m *mockCache) InvalidateAll() error {
	m.InvalidateAllCalled = true
	return nil
}

func (m *mockCache) RefreshFlags() error {
	m.RefreshCalled = true
	return nil
}

// ----------------------
// AdminServer Tests
// ----------------------

func TestAdminServer_Health(t *testing.T) {
	srv := NewAdminServer(&mockCache{}, 0)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "healthy", resp["status"])
}

func TestAdminServer_Stats(t *testing.T) {
	mock := &mockCache{Metrics: map[string]int{"a": 1}}
	srv := NewAdminServer(mock, 0)

	req := httptest.NewRequest("GET", "/admin/stats", nil)
	w := httptest.NewRecorder()

	srv.handleStats(w, req)
	assert.Equal(t, 200, w.Code)

	var resp map[string]int
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, 1, resp["a"])
}

func TestAdminServer_Invalidate(t *testing.T) {
	mock := &mockCache{}
	srv := NewAdminServer(mock, 0)

	body := bytes.NewBufferString(`{"flag_key":"test_flag"}`)
	req := httptest.NewRequest("POST", "/admin/invalidate", body)
	w := httptest.NewRecorder()

	srv.handleInvalidate(w, req)

	assert.True(t, mock.InvalidateCalled)
	assert.Equal(t, "test_flag", mock.LastInvalidatedKey)
}

func TestAdminServer_InvalidateAll(t *testing.T) {
	mock := &mockCache{}
	srv := NewAdminServer(mock, 0)

	req := httptest.NewRequest("POST", "/admin/invalidate-all", nil)
	w := httptest.NewRecorder()

	srv.handleInvalidateAll(w, req)

	assert.True(t, mock.InvalidateAllCalled)
}

func TestAdminServer_Refresh(t *testing.T) {
	mock := &mockCache{}
	srv := NewAdminServer(mock, 0)

	req := httptest.NewRequest("POST", "/admin/refresh", nil)
	w := httptest.NewRecorder()

	srv.handleRefresh(w, req)

	assert.True(t, mock.RefreshCalled)
}

// ----------------------
// Middleware Tests
// ----------------------

func TestMiddleware_Handler(t *testing.T) {
	mock := &mockCache{}
	mw := NewMiddleware(mock)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		eval, ok := GetEvalContext(r.Context())
		assert.True(t, ok)
		assert.NotEmpty(t, eval["method"])
		assert.NotEmpty(t, eval["path"])

		cache, ok := GetCache(r.Context())
		assert.True(t, ok)
		assert.Equal(t, mock, cache)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mw.Handler(next).ServeHTTP(w, req)
	assert.True(t, called)
}

// ----------------------
// Webhook Tests
// ----------------------

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
