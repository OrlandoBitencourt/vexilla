package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAdminHandler struct {
	stats           interface{}
	invalidatedKeys []string
	invalidatedAll  bool
	refreshed       bool
}

func (m *mockAdminHandler) GetStats(ctx context.Context) (interface{}, error) {
	return m.stats, nil
}

func (m *mockAdminHandler) InvalidateFlags(ctx context.Context, flagKeys []string) error {
	m.invalidatedKeys = append(m.invalidatedKeys, flagKeys...)
	return nil
}

func (m *mockAdminHandler) InvalidateAll(ctx context.Context) error {
	m.invalidatedAll = true
	return nil
}

func (m *mockAdminHandler) ForceRefresh(ctx context.Context) error {
	m.refreshed = true
	return nil
}

func (m *mockAdminHandler) HealthCheck(ctx context.Context) (interface{}, error) {
	return map[string]string{"status": "healthy"}, nil
}

func TestAdminServer_GetStats(t *testing.T) {
	handler := &mockAdminHandler{
		stats: map[string]interface{}{
			"cache_hits":   100,
			"cache_misses": 10,
		},
	}

	server := NewAdminServer(8082, "/admin", handler)

	req := httptest.NewRequest("GET", "/admin/stats", nil)
	rw := httptest.NewRecorder()

	server.handleStats(rw, req)

	assert.Equal(t, 200, rw.Code)

	var response map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&response)
	assert.Equal(t, float64(100), response["cache_hits"])
}

func TestAdminServer_Invalidate(t *testing.T) {
	handler := &mockAdminHandler{}
	server := NewAdminServer(8082, "/admin", handler)

	reqBody := map[string]interface{}{
		"flag_keys": []string{"flag1", "flag2"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/invalidate", bytes.NewReader(body))
	rw := httptest.NewRecorder()

	server.handleInvalidate(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, []string{"flag1", "flag2"}, handler.invalidatedKeys)
}

func TestAdminServer_InvalidateAll(t *testing.T) {
	handler := &mockAdminHandler{}
	server := NewAdminServer(8082, "/admin", handler)

	req := httptest.NewRequest("POST", "/admin/invalidate-all", nil)
	rw := httptest.NewRecorder()

	server.handleInvalidateAll(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.True(t, handler.invalidatedAll)
}

func TestAdminServer_Refresh(t *testing.T) {
	handler := &mockAdminHandler{}
	server := NewAdminServer(8082, "/admin", handler)

	req := httptest.NewRequest("POST", "/admin/refresh", nil)
	rw := httptest.NewRecorder()

	server.handleRefresh(rw, req)

	assert.Equal(t, 200, rw.Code)
	assert.True(t, handler.refreshed)
}
