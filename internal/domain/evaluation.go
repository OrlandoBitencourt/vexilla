package domain

import (
	"encoding/json"
	"time"
)

// EvaluationContext holds user/request context for flag evaluation
type EvaluationContext struct {
	// EntityID is the unique identifier for the entity (user, account, etc.)
	EntityID string

	// EntityType describes the type of entity (user, account, device, etc.)
	EntityType string

	// Context contains additional attributes for constraint evaluation
	Context map[string]interface{}
}

// EvaluationResult represents the result of flag evaluation
type EvaluationResult struct {
	// Flag information
	FlagID  int64
	FlagKey string

	// Segment information
	SegmentID int64

	// Variant information
	VariantID         int64
	VariantKey        string
	VariantAttachment map[string]json.RawMessage

	// Evaluation metadata
	EvaluationReason string
	Timestamp        time.Time

	// Performance metrics
	EvaluationTime time.Duration
}

// NewEvaluationContext creates a new evaluation context
func NewEvaluationContext(entityID string) EvaluationContext {
	return EvaluationContext{
		EntityID:   entityID,
		EntityType: "user",
		Context:    make(map[string]interface{}),
	}
}

// WithAttribute adds an attribute to the context
func (e EvaluationContext) WithAttribute(key string, value interface{}) EvaluationContext {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithAttributes adds multiple attributes to the context
func (e EvaluationContext) WithAttributes(attrs map[string]interface{}) EvaluationContext {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range attrs {
		e.Context[k] = v
	}
	return e
}

// WithEntityType sets the entity type
func (e EvaluationContext) WithEntityType(entityType string) EvaluationContext {
	e.EntityType = entityType
	return e
}

// GetAttribute retrieves an attribute from the context
func (e EvaluationContext) GetAttribute(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	val, ok := e.Context[key]
	return val, ok
}

// HasAttribute checks if an attribute exists
func (e EvaluationContext) HasAttribute(key string) bool {
	_, ok := e.GetAttribute(key)
	return ok
}

// ----------------------------
// Helpers to decode RawMessage
// ----------------------------

func decodeBool(m map[string]json.RawMessage, key string) (bool, bool) {
	raw, ok := m[key]
	if !ok {
		return false, false
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err == nil {
		return v, true
	}
	return false, false
}

func decodeString(m map[string]json.RawMessage, key string) (string, bool) {
	raw, ok := m[key]
	if !ok {
		return "", false
	}
	var v string
	if err := json.Unmarshal(raw, &v); err == nil {
		return v, true
	}
	return "", false
}

func decodeInt(m map[string]json.RawMessage, key string) (int, bool) {
	raw, ok := m[key]
	if !ok {
		return 0, false
	}
	var v float64
	if err := json.Unmarshal(raw, &v); err == nil {
		return int(v), true
	}
	return 0, false
}

func decodeFloat(m map[string]json.RawMessage, key string) (float64, bool) {
	raw, ok := m[key]
	if !ok {
		return 0, false
	}
	var v float64
	if err := json.Unmarshal(raw, &v); err == nil {
		return v, true
	}
	return 0, false
}

// IsEnabled returns true if the result indicates an enabled feature
func (r *EvaluationResult) IsEnabled() bool {
	if r.VariantAttachment == nil {
		return r.VariantKey == "enabled" ||
			r.VariantKey == "on" ||
			r.VariantKey == "true"
	}

	if v, ok := decodeBool(r.VariantAttachment, "enabled"); ok {
		return v
	}
	if v, ok := decodeBool(r.VariantAttachment, "value"); ok {
		return v
	}

	return false
}

// GetStringValue returns the string value from the result
func (r *EvaluationResult) GetStringValue(defaultVal string) string {
	if r.VariantAttachment == nil {
		if r.VariantKey != "" {
			return r.VariantKey
		}
		return defaultVal
	}

	if v, ok := decodeString(r.VariantAttachment, "value"); ok {
		return v
	}

	return defaultVal
}

// GetIntValue returns the int value
func (r *EvaluationResult) GetIntValue(defaultVal int) int {
	if r.VariantAttachment == nil {
		return defaultVal
	}

	if v, ok := decodeInt(r.VariantAttachment, "value"); ok {
		return v
	}

	return defaultVal
}

// GetFloatValue returns the float value
func (r *EvaluationResult) GetFloatValue(defaultVal float64) float64 {
	if r.VariantAttachment == nil {
		return defaultVal
	}

	if v, ok := decodeFloat(r.VariantAttachment, "value"); ok {
		return v
	}

	return defaultVal
}

// GetJSONValue returns a JSON-decoded value for a key
func (r *EvaluationResult) GetJSONValue(key string) interface{} {
	if r.VariantAttachment == nil {
		return nil
	}

	raw, ok := r.VariantAttachment[key]
	if !ok {
		return nil
	}

	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}
