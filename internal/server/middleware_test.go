package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
