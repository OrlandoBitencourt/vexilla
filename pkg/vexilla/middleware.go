package vexilla

import (
	"context"
	"fmt"
	"net/http"
)

type contextKey string

const (
	contextKeyEvalCtx contextKey = "vexilla_eval_ctx"
	contextKeyClient  contextKey = "vexilla_client"
)

// Middleware returns an HTTP middleware that injects Vexilla client into context
func (c *Client) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := c.tracer.Tracer().Start(r.Context(), "vexilla.middleware")
		defer span.End()

		// Build evaluation context from request
		evalCtx := c.buildEvaluationContext(r)

		// Inject into context
		ctx = context.WithValue(ctx, contextKeyEvalCtx, evalCtx)
		ctx = context.WithValue(ctx, contextKeyClient, c)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// buildEvaluationContext creates evaluation context from HTTP request
func (c *Client) buildEvaluationContext(r *http.Request) EvaluationContext {
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
		EntityID: userID,
		Context:  attributes,
	}
}

// GetFlagFromContext evaluates a flag from request context
func GetFlagFromContext(ctx context.Context, flagKey string) (*EvaluationResult, error) {
	client, ok := ctx.Value(contextKeyClient).(*Client)
	if !ok {
		return nil, fmt.Errorf("vexilla client not found in context")
	}

	evalCtx, ok := ctx.Value(contextKeyEvalCtx).(EvaluationContext)
	if !ok {
		evalCtx = EvaluationContext{}
	}

	return client.Evaluate(ctx, flagKey, evalCtx)
}

// GetFlagBoolFromContext is a convenience helper for boolean flags
func GetFlagBoolFromContext(ctx context.Context, flagKey string) bool {
	result, err := GetFlagFromContext(ctx, flagKey)
	if err != nil {
		return false
	}
	return result.VariantKey == "enabled" || result.VariantKey == "on"
}
