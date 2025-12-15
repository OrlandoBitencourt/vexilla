package flagr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create RawMessage
func raw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestNewHTTPClient(t *testing.T) {
	config := Config{
		Endpoint:   "http://localhost:18000",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	}

	client := NewHTTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, config.Endpoint, client.endpoint)
	assert.Equal(t, config.MaxRetries, client.maxRetries)
}

func TestHTTPClient_GetFlag(t *testing.T) {
	mockFlag := FlagrFlag{
		ID:          1,
		Key:         "test-flag",
		Description: "Test flag",
		Enabled:     true,
		Segments: []FlagrSegment{
			{
				ID:             1,
				RolloutPercent: 100,
				Distributions: []FlagrDistribution{
					{ID: 1, VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []FlagrVariant{
			{ID: 1, Key: "on", Attachment: map[string]json.RawMessage{"enabled": raw(true)}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/flags/1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockFlag)
	}))
	defer server.Close()

	client := NewHTTPClient(Config{
		Endpoint:   server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	ctx := context.Background()
	flag, err := client.GetFlag(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), flag.ID)
	assert.Equal(t, "test-flag", flag.Key)
	assert.True(t, flag.Enabled)
}

func TestHTTPClient_EvaluateFlag(t *testing.T) {
	mockResponse := EvaluationResponse{
		FlagID:     1,
		FlagKey:    "test-flag",
		SegmentID:  1,
		VariantID:  1,
		VariantKey: "on",
		VariantAttachment: map[string]json.RawMessage{
			"enabled": raw(true),
		},
		Timestamp: time.Now(),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/evaluation", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req EvaluationRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-flag", req.FlagKey)
		assert.Equal(t, "user123", req.EntityID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewHTTPClient(Config{
		Endpoint:   server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context:  map[string]interface{}{"country": "BR"},
	}

	result, err := client.EvaluateFlag(ctx, "test-flag", evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "test-flag", result.FlagKey)
	assert.Equal(t, "on", result.VariantKey)
	assert.Equal(t, raw(true), result.VariantAttachment["enabled"])
}

func TestHTTPClient_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   HealthResponse
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "healthy",
			mockResponse:   HealthResponse{Status: "OK"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "unhealthy status",
			mockResponse:   HealthResponse{Status: "ERROR"},
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
		{
			name:           "server error",
			mockResponse:   HealthResponse{},
			mockStatusCode: http.StatusServiceUnavailable,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/health", r.URL.Path)
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := NewHTTPClient(Config{
				Endpoint:   server.URL,
				Timeout:    5 * time.Second,
				MaxRetries: 0,
			})

			ctx := context.Background()
			err := client.HealthCheck(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPClient_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]FlagrFlag{})
	}))
	defer server.Close()

	client := NewHTTPClient(Config{
		Endpoint:   server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})

	ctx := context.Background()
	_, err := client.GetAllFlags(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestHTTPClient_Authentication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer secret-token", auth)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]FlagrFlag{})
	}))
	defer server.Close()

	client := NewHTTPClient(Config{
		Endpoint:   server.URL,
		APIKey:     "secret-token",
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	ctx := context.Background()
	_, err := client.GetAllFlags(ctx)

	assert.NoError(t, err)
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(Config{
		Endpoint:   server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetAllFlags(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

// helper to create client pointing to mock server
func newTestClient(serverURL string) *HTTPClient {
	return &HTTPClient{
		endpoint: serverURL,
		apiKey:   "testkey",
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		maxRetries: 0, // avoid retry in tests unless testing retry behavior
	}
}

//
// ────────────────────────────────────────────────
//   Test: GetFlag
// ────────────────────────────────────────────────
//

func TestGetFlag_Success(t *testing.T) {
	mockFlag := FlagrFlag{
		ID:          10,
		Key:         "my_flag",
		Description: "test flag",
		Enabled:     true,
		Tags:        []Tag{{Value: "service:test"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockFlag)
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	res, err := client.GetFlag(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != 10 || res.Key != "my_flag" {
		t.Fatalf("flag not correctly parsed: %+v", res)
	}
}

func TestGetFlag_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	_, err := client.GetFlag(context.Background(), 99)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

//
// ────────────────────────────────────────────────
//   Test: GetAllFlags
// ────────────────────────────────────────────────
//

func TestGetAllFlags_Success(t *testing.T) {
	flagList := []FlagrFlag{
		{ID: 1, Key: "flag1"},
		{ID: 2, Key: "flag2"},
	}

	// track which endpoint is called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch path {
		case "/api/v1/flags":
			json.NewEncoder(w).Encode(flagList)
		case "/api/v1/flags/1":
			json.NewEncoder(w).Encode(flagList[0])
		case "/api/v1/flags/2":
			json.NewEncoder(w).Encode(flagList[1])
		default:
			t.Fatalf("unexpected path: %s", path)
		}
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	flags, err := client.GetAllFlags(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
}

//
// ────────────────────────────────────────────────
//   Test: EvaluateFlag
// ────────────────────────────────────────────────
//

func TestEvaluateFlag_Success(t *testing.T) {
	mockEval := EvaluationResponse{
		FlagID:     1,
		FlagKey:    "my_flag",
		SegmentID:  10,
		VariantID:  100,
		VariantKey: "enabled",
		VariantAttachment: map[string]json.RawMessage{
			"enabled": json.RawMessage(`true`),
		},
		Timestamp:    time.Now(),
		EvalDebugLog: EvalDebugLog{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockEval)
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	ctx := domain.NewEvaluationContext("user-1")

	res, err := client.EvaluateFlag(context.Background(), "my_flag", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsEnabled() {
		t.Fatalf("expected enabled flag")
	}
}

//
// ────────────────────────────────────────────────
//   Test: HealthCheck
// ────────────────────────────────────────────────
//

func TestHealthCheck_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{Status: "OK"})
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	if err := client.HealthCheck(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealthCheck_Fail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatalf("expected failure but got nil")
	}
}

//
// ────────────────────────────────────────────────
//   Test: doRequest retry logic
// ────────────────────────────────────────────────
//

func TestDoRequest_RetryOn5xx(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "retry pls", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.maxRetries = 5

	var result map[string]string
	err := client.doRequest(context.Background(), "GET", server.URL, nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoRequest_NoRetryOn400(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	client.maxRetries = 3

	var result map[string]string
	err := client.doRequest(context.Background(), "GET", server.URL, nil, &result)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if attempts != 1 {
		t.Fatalf("should not retry 400 errors")
	}
}

//
// ────────────────────────────────────────────────
//   Test: shouldRetry
// ────────────────────────────────────────────────
//

func TestShouldRetry(t *testing.T) {
	client := newTestClient("")

	// Retry on 500
	err500 := &HTTPError{StatusCode: 500}
	if !client.shouldRetry(err500) {
		t.Fatalf("should retry on 500")
	}

	// Retry on 429
	err429 := &HTTPError{StatusCode: 429}
	if !client.shouldRetry(err429) {
		t.Fatalf("should retry on 429")
	}

	// No retry on 400
	err400 := &HTTPError{StatusCode: 400}
	if client.shouldRetry(err400) {
		t.Fatalf("should not retry on 400")
	}

	// Network errors should retry
	if !client.shouldRetry(errors.New("network")) {
		t.Fatalf("should retry network errors")
	}
}
