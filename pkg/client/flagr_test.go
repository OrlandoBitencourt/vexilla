package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagrClient_GetFlags(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/flags", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		flags := []vexilla.Flag{
			{ID: 1, Key: "flag1", Enabled: true},
			{ID: 2, Key: "flag2", Enabled: false},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flags)
	}))
	defer server.Close()

	client := NewFlagrClient(server.URL, "", 5*time.Second)
	ctx := context.Background()

	flags, err := client.GetFlags(ctx)
	require.NoError(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "flag1", flags[0].Key)
	assert.Equal(t, "flag2", flags[1].Key)
}

func TestFlagrClient_PostEvaluation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/evaluation", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		assert.Equal(t, "test_flag", req["flagKey"])
		assert.Equal(t, "user123", req["entityID"])

		result := map[string]interface{}{
			"flagID":     1,
			"flagKey":    "test_flag",
			"segmentID":  10,
			"variantID":  100,
			"variantKey": "enabled",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewFlagrClient(server.URL, "", 5*time.Second)
	ctx := context.Background()

	evalCtx := vexilla.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "US",
		},
	}

	result, err := client.PostEvaluation(ctx, "test_flag", evalCtx)
	require.NoError(t, err)
	assert.Equal(t, "test_flag", result.FlagKey)
	assert.Equal(t, "enabled", result.VariantKey)
	assert.False(t, result.EvaluatedLocally)
}

func TestFlagrClient_Health(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewFlagrClient(server.URL, "", 5*time.Second)
	ctx := context.Background()

	err := client.Health(ctx)
	assert.NoError(t, err)
}

func TestFlagrClient_GetFlags_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewFlagrClient(server.URL, "", 5*time.Second)
	ctx := context.Background()

	_, err := client.GetFlags(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
