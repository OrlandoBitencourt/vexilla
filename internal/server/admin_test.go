package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAdminServer_InvalidateRequestBodyTooLarge(t *testing.T) {
	mock := &mockCache{}
	srv := NewAdminServer(mock, 0)

	// Create a valid JSON payload larger than MaxRequestBodySize (1MB)
	// Build a large flag_key value
	largeFlagKey := make([]byte, MaxRequestBodySize+100)
	for i := range largeFlagKey {
		largeFlagKey[i] = 'A'
	}

	payload := map[string]string{
		"flag_key": string(largeFlagKey),
	}
	largeJSON, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/admin/invalidate", bytes.NewBuffer(largeJSON))
	w := httptest.NewRecorder()

	srv.handleInvalidate(w, req)

	// Should return 413 Request Entity Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.False(t, mock.InvalidateCalled)
}

func TestAdminServer_InvalidateRequestBodyAtLimit(t *testing.T) {
	mock := &mockCache{}
	srv := NewAdminServer(mock, 0)

	// Create a valid JSON payload just under the size limit
	// Build a large flag key
	largeFlagKeyBytes := make([]byte, MaxRequestBodySize-100)
	for i := range largeFlagKeyBytes {
		largeFlagKeyBytes[i] = 'a'
	}

	payload := map[string]string{
		"flag_key": string(largeFlagKeyBytes),
	}
	body, _ := json.Marshal(payload)

	// Verify the body is under the limit
	if len(body) > MaxRequestBodySize {
		// If still too large, use a smaller payload
		payload["flag_key"] = "test_flag"
		body, _ = json.Marshal(payload)
	}

	req := httptest.NewRequest("POST", "/admin/invalidate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	srv.handleInvalidate(w, req)

	// Should succeed with payload under the limit
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, mock.InvalidateCalled)
}

func TestAdminServer_StartShutdown(t *testing.T) {
	mock := &mockCache{Metrics: map[string]string{"test": "value"}}
	admin := NewAdminServer(mock, 19999)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- admin.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is running
	resp, err := http.Get("http://localhost:19999/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should get success
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	err = admin.Shutdown(shutdownCtx)
	assert.NoError(t, err)

	// Wait for Start to return
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after shutdown")
	}
}

func TestAdminServer_Shutdown_NotStarted(t *testing.T) {
	mock := &mockCache{}
	admin := NewAdminServer(mock, 0)

	// Shutdown without starting should not error
	err := admin.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestAdminServer_ContextCancellation(t *testing.T) {
	mock := &mockCache{}
	admin := NewAdminServer(mock, 19998)

	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- admin.Start(ctx)
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
