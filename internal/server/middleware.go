package server

import (
	"context"
	"net/http"
)

type contextKey string

const (
	contextKeyEvalCtx contextKey = "vexilla_eval_ctx"
	contextKeyCache   contextKey = "vexilla_cache"
)

// Middleware provides HTTP middleware for flag injection
type Middleware struct {
	cache CacheInterface
}

// NewMiddleware creates new middleware
func NewMiddleware(cache CacheInterface) *Middleware {
	return &Middleware{cache: cache}
}

// Handler wraps an HTTP handler with flag evaluation context
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := m.buildContext(r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) buildContext(r *http.Request) context.Context {
	ctx := r.Context()

	// Extract user ID from header or cookie
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		if cookie, err := r.Cookie("user_id"); err == nil {
			userID = cookie.Value
		}
	}

	// Build evaluation context
	evalCtx := map[string]interface{}{
		"user_id":    userID,
		"ip":         r.RemoteAddr,
		"user_agent": r.UserAgent(),
		"path":       r.URL.Path,
		"method":     r.Method,
	}

	// Add headers as attributes
	for key, values := range r.Header {
		if len(values) > 0 {
			evalCtx["header_"+key] = values[0]
		}
	}

	ctx = context.WithValue(ctx, contextKeyEvalCtx, evalCtx)
	ctx = context.WithValue(ctx, contextKeyCache, m.cache)

	return ctx
}

// GetEvalContext extracts evaluation context from request context
func GetEvalContext(ctx context.Context) (map[string]interface{}, bool) {
	evalCtx, ok := ctx.Value(contextKeyEvalCtx).(map[string]interface{})
	return evalCtx, ok
}

// GetCache extracts cache from request context
func GetCache(ctx context.Context) (CacheInterface, bool) {
	cache, ok := ctx.Value(contextKeyCache).(CacheInterface)
	return cache, ok
}
