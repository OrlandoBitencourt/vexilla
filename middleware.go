package vexilla

import (
	"context"
	"fmt"
	"net/http"
)

type contextKey string

const (
	contextKeyEvalCtx contextKey = "vexilla_eval_ctx"
	contextKeyCache   contextKey = "vexilla_cache"
)

// Middleware returns an HTTP middleware that injects flag evaluations
func (c *Cache) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := c.tracer.Start(r.Context(), "middleware.flag_injection")
		defer span.End()

		evalCtx := c.buildEvaluationContext(r)
		ctx = context.WithValue(ctx, contextKeyEvalCtx, evalCtx)
		ctx = context.WithValue(ctx, contextKeyCache, c)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// buildEvaluationContext creates an evaluation context from HTTP request
func (c *Cache) buildEvaluationContext(r *http.Request) EvaluationContext {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		if cookie, err := r.Cookie("user_id"); err == nil {
			userID = cookie.Value
		}
	}

	attributes := map[string]interface{}{
		"ip":         r.RemoteAddr,
		"user_agent": r.UserAgent(),
		"path":       r.URL.Path,
		"method":     r.Method,
	}

	for key, values := range r.Header {
		if len(values) > 0 {
			attributes["header_"+key] = values[0]
		}
	}

	return EvaluationContext{
		UserID:     userID,
		Attributes: attributes,
	}
}

// GetFlagFromContext evaluates a flag from request context
func GetFlagFromContext(ctx context.Context, flagKey string) (interface{}, error) {
	cache, ok := ctx.Value(contextKeyCache).(*Cache)
	if !ok {
		return nil, fmt.Errorf("vexilla cache not found in context")
	}

	evalCtx, ok := ctx.Value(contextKeyEvalCtx).(EvaluationContext)
	if !ok {
		evalCtx = EvaluationContext{}
	}

	return cache.Evaluate(ctx, flagKey, evalCtx)
}

// GetFlagBoolFromContext is a convenience helper
func GetFlagBoolFromContext(ctx context.Context, flagKey string) bool {
	result, err := GetFlagFromContext(ctx, flagKey)
	if err != nil {
		return false
	}
	if b, ok := result.(bool); ok {
		return b
	}
	return false
}
