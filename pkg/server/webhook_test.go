package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWebhookHandler struct {
	updatedKeys []string
	deletedKeys []string
}

func (m *mockWebhookHandler) OnFlagUpdated(ctx context.Context, flagKeys []string) error {
	m.updatedKeys = append(m.updatedKeys, flagKeys...)
	return nil
}

func (m *mockWebhookHandler) OnFlagDeleted(ctx context.Context, flagKeys []string) error {
	m.deletedKeys = append(m.deletedKeys, flagKeys...)
	return nil
}

func TestWebhookServer_HandleFlagUpdated(t *testing.T) {
	handler := &mockWebhookHandler{}
	server, err := NewWebhookServer(8081, "/webhook", "", handler)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:    "flag.updated",
		FlagKeys: []string{"flag1", "flag2"},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	rw := httptest.NewRecorder()

	server.handleWebhook(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, []string{"flag1", "flag2"}, handler.updatedKeys)
}

func TestWebhookServer_HandleFlagDeleted(t *testing.T) {
	handler := &mockWebhookHandler{}
	server, err := NewWebhookServer(8081, "/webhook", "", handler)
	require.NoError(t, err)

	payload := WebhookPayload{
		Event:    "flag.deleted",
		FlagKeys: []string{"flag3"},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	rw := httptest.NewRecorder()

	server.handleWebhook(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, []string{"flag3"}, handler.deletedKeys)
}

func TestWebhookServer_SecretValidation(t *testing.T) {
	handler := &mockWebhookHandler{}
	server, err := NewWebhookServer(8081, "/webhook", "secret123", handler)
	require.NoError(t, err)

	payload := WebhookPayload{Event: "flag.updated", FlagKeys: []string{"flag1"}}
	body, _ := json.Marshal(payload)

	// Without secret - should fail
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	rw := httptest.NewRecorder()
	server.handleWebhook(rw, req)
	assert.Equal(t, http.StatusUnauthorized, rw.Code)

	// With correct secret - should succeed
	req = httptest.NewRequest("POST", "/webhook", bytes.NewReader(body))
	req.Header.Set("X-Webhook-Secret", "secret123")
	rw = httptest.NewRecorder()
	server.handleWebhook(rw, req)
	assert.Equal(t, http.StatusOK, rw.Code)
}
