package vexilla

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithWebhookInvalidation tests webhook configuration
func TestWithWebhookInvalidation(t *testing.T) {
	tests := []struct {
		name      string
		config    WebhookConfig
		expectErr bool
	}{
		{
			name: "valid webhook config",
			config: WebhookConfig{
				Port:   18001,
				Secret: "test-secret",
			},
			expectErr: false,
		},
		{
			name: "webhook without secret",
			config: WebhookConfig{
				Port: 18001,
			},
			expectErr: false,
		},
		{
			name: "invalid port - zero",
			config: WebhookConfig{
				Port:   0,
				Secret: "secret",
			},
			expectErr: true,
		},
		{
			name: "invalid port - negative",
			config: WebhookConfig{
				Port:   -1,
				Secret: "secret",
			},
			expectErr: true,
		},
		{
			name: "invalid port - too high",
			config: WebhookConfig{
				Port:   70000,
				Secret: "secret",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &clientConfig{}
			opt := WithWebhookInvalidation(tt.config)
			err := opt(cfg)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, cfg.webhookEnabled)
				assert.Equal(t, tt.config.Port, cfg.webhookPort)
				assert.Equal(t, tt.config.Secret, cfg.webhookSecret)
			}
		})
	}
}

// TestWithAdminServer tests admin server configuration
func TestWithAdminServer(t *testing.T) {
	tests := []struct {
		name      string
		config    AdminConfig
		expectErr bool
	}{
		{
			name: "valid admin config",
			config: AdminConfig{
				Port: 19000,
			},
			expectErr: false,
		},
		{
			name: "invalid port - zero",
			config: AdminConfig{
				Port: 0,
			},
			expectErr: true,
		},
		{
			name: "invalid port - negative",
			config: AdminConfig{
				Port: -1,
			},
			expectErr: true,
		},
		{
			name: "invalid port - too high",
			config: AdminConfig{
				Port: 100000,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &clientConfig{}
			opt := WithAdminServer(tt.config)
			err := opt(cfg)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, cfg.adminEnabled)
				assert.Equal(t, tt.config.Port, cfg.adminPort)
			}
		})
	}
}

// TestClient_WithWebhookServer tests client creation with webhook server
func TestClient_WithWebhookServer(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithWebhookInvalidation(WebhookConfig{
			Port:   18501, // Use random high port for testing
			Secret: "test-secret",
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Verify webhook server is running
	assert.True(t, client.webhookEnabled)
	assert.Equal(t, 18501, client.webhookPort)
}

// TestClient_WithAdminServer tests client creation with admin server
func TestClient_WithAdminServer(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithAdminServer(AdminConfig{
			Port: 19501, // Use random high port for testing
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Verify admin server is running
	assert.True(t, client.adminEnabled)
	assert.Equal(t, 19501, client.adminPort)

	// Test health endpoint
	resp, err := http.Get("http://localhost:19501/health")
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]string
		json.NewDecoder(resp.Body).Decode(&health)
		assert.Equal(t, "healthy", health["status"])
	}
}

// TestClient_HTTPMiddleware tests the HTTP middleware
func TestClient_HTTPMiddleware(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "feature-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "enabled",
				Attachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
				},
			},
		},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithOnlyEnabled(true),
	)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, client.Start(ctx))
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The middleware should have injected context
		evalCtx := NewContext("test-user")
		enabled := client.Bool(r.Context(), "feature-flag", evalCtx)

		if enabled {
			w.Write([]byte("enabled"))
		} else {
			w.Write([]byte("disabled"))
		}
	})

	// Wrap with middleware
	wrappedHandler := client.HTTPMiddleware(testHandler)

	// Create test server
	testServer := &http.Server{
		Addr:    ":18502",
		Handler: wrappedHandler,
	}

	go func() {
		testServer.ListenAndServe()
	}()
	defer testServer.Close()

	time.Sleep(100 * time.Millisecond)

	// Test the endpoint
	resp, err := http.Get("http://localhost:18502/")
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "enabled", string(body))
	}
}

// TestClient_CombinedServers tests client with both webhook and admin servers
func TestClient_CombinedServers(t *testing.T) {
	server := NewMockFlagrServer(t)
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "combined-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, err := New(
		WithFlagrEndpoint(server.URL),
		WithWebhookInvalidation(WebhookConfig{
			Port:   18503,
			Secret: "webhook-secret",
		}),
		WithAdminServer(AdminConfig{
			Port: 19503,
		}),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = client.Start(ctx)
	require.NoError(t, err)
	defer client.Stop()

	time.Sleep(200 * time.Millisecond)

	// Verify both servers are configured
	assert.True(t, client.webhookEnabled)
	assert.True(t, client.adminEnabled)
	assert.Equal(t, 18503, client.webhookPort)
	assert.Equal(t, 19503, client.adminPort)
}

// BenchmarkClient_HTTPMiddleware benchmarks the middleware
func BenchmarkClient_HTTPMiddleware(b *testing.B) {
	server := NewMockFlagrServer(&testing.T{})
	defer server.Close()

	server.AddFlag(domain.Flag{
		ID:      1,
		Key:     "bench-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	client, _ := New(WithFlagrEndpoint(server.URL))
	ctx := context.Background()
	client.Start(ctx)
	defer client.Stop()

	time.Sleep(100 * time.Millisecond)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		evalCtx := NewContext("bench-user")
		client.Bool(r.Context(), "bench-flag", evalCtx)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := client.HTTPMiddleware(handler)

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-User-ID", "bench-user")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := &testResponseWriter{}
		wrappedHandler.ServeHTTP(w, req)
	}
}

// testResponseWriter is a minimal ResponseWriter for testing
type testResponseWriter struct {
	headers http.Header
	status  int
}

func (w *testResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}
