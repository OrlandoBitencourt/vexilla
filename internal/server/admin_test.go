package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCache struct {
	InvalidateCalled    bool
	InvalidateAllCalled bool
	RefreshCalled       bool
	LastInvalidatedKey  string
	Metrics             interface{}
	invalidateErr       error
	invalidateAllErr    error
	refreshErr          error
}

func (m *mockCache) GetMetrics() interface{} {
	return m.Metrics
}

func (m *mockCache) InvalidateFlag(flagKey string) error {
	m.InvalidateCalled = true
	m.LastInvalidatedKey = flagKey
	return m.invalidateErr
}

func (m *mockCache) InvalidateAll() error {
	m.InvalidateAllCalled = true
	return m.invalidateAllErr
}

func (m *mockCache) RefreshFlags() error {
	m.RefreshCalled = true
	return m.refreshErr
}

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
