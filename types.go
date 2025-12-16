package vexilla

import (
	"encoding/json"
	"strings"
	"time"
)

// Context holds user/request context for flag evaluation.
type Context struct {
	// EntityID is the unique identifier for the entity being evaluated
	// (e.g., user ID, account ID, device ID)
	EntityID string

	// EntityType describes the type of entity (default: "user")
	EntityType string

	// Attributes contains additional context for constraint evaluation
	// (e.g., country, tier, age, etc.)
	Attributes map[string]any
}

// NewContext creates a new evaluation context with the given entity ID.
func NewContext(entityID string) Context {
	return Context{
		EntityID:   entityID,
		EntityType: "user",
		Attributes: make(map[string]any),
	}
}

// WithAttribute adds an attribute to the context (fluent interface).
func (c Context) WithAttribute(key string, value any) Context {
	if c.Attributes == nil {
		c.Attributes = make(map[string]any)
	}
	c.Attributes[key] = value
	return c
}

// WithEntityType sets the entity type (fluent interface).
func (c Context) WithEntityType(entityType string) Context {
	c.EntityType = entityType
	return c
}

// Result represents the result of flag evaluation.
type Result struct {
	// FlagKey is the key of the evaluated flag
	FlagKey string

	// VariantKey is the key of the matched variant
	VariantKey string

	// VariantAttachment contains the variant's JSON data
	VariantAttachment map[string]json.RawMessage

	// EvaluationReason explains why this result was returned
	EvaluationReason string
}

// IsEnabled returns true if the result indicates an enabled feature.
func (r *Result) IsEnabled() bool {
	if r.VariantAttachment == nil {
		return r.VariantKey == "enabled" ||
			r.VariantKey == "on" ||
			r.VariantKey == "true"
	}

	if isTrue(r.VariantAttachment["enabled"]) {
		return true
	}

	if isTrue(r.VariantAttachment["value"]) {
		return true
	}

	return false
}

func isTrue(raw json.RawMessage) bool {
	if raw == nil {
		return false
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		switch strings.ToLower(s) {
		case "true", "on", "enabled", "1", "yes":
			return true
		}
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n == 1
	}

	return false
}

// GetString returns a string value from the variant attachment.
func (r *Result) GetString(key string, defaultVal string) string {
	if r.VariantAttachment == nil {
		return defaultVal
	}

	raw, ok := r.VariantAttachment[key]
	if !ok {
		return defaultVal
	}

	var v string
	if err := json.Unmarshal(raw, &v); err != nil {
		return defaultVal
	}

	return v
}

// GetInt returns an integer value from the variant attachment.
func (r *Result) GetInt(key string, defaultVal int) int {
	if r.VariantAttachment == nil {
		return defaultVal
	}

	raw, ok := r.VariantAttachment[key]
	if !ok {
		return defaultVal
	}

	var v int
	if err := json.Unmarshal(raw, &v); err == nil {
		return v
	}

	// Try float64 (JSON default)
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int(f)
	}

	return defaultVal
}

// Metrics represents cache performance metrics.
type Metrics struct {
	// Storage metrics
	Storage StorageMetrics

	// LastRefresh is when the cache was last refreshed
	LastRefresh time.Time

	// ConsecutiveFails is the number of consecutive refresh failures
	ConsecutiveFails int

	// CircuitOpen indicates if the circuit breaker is open
	CircuitOpen bool
}

// StorageMetrics represents storage layer metrics.
type StorageMetrics struct {
	// KeysAdded is the total number of keys added to cache
	KeysAdded uint64

	// KeysEvicted is the total number of keys evicted
	KeysEvicted uint64

	// HitRatio is the cache hit ratio (0.0 to 1.0)
	HitRatio float64
}
