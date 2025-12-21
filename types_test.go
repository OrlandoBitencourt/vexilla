package vexilla

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContext_WithAttribute tests adding attributes to context
func TestContext_WithAttribute(t *testing.T) {
	// Test with nil attributes map
	ctx := Context{
		EntityID: "user-123",
	}
	ctx = ctx.WithAttribute("country", "BR")
	assert.Equal(t, "BR", ctx.Attributes["country"])

	// Test with existing attributes map
	ctx2 := NewContext("user-456")
	ctx2 = ctx2.WithAttribute("tier", "premium")
	ctx2 = ctx2.WithAttribute("age", 25)

	assert.Equal(t, "premium", ctx2.Attributes["tier"])
	assert.Equal(t, 25, ctx2.Attributes["age"])
}

// TestContext_WithEntityType tests setting entity type
func TestContext_WithEntityType(t *testing.T) {
	ctx := NewContext("device-789")
	ctx = ctx.WithEntityType("device")

	assert.Equal(t, "device-789", ctx.EntityID)
	assert.Equal(t, "device", ctx.EntityType)
}

// TestResult_IsEnabled tests various enabled/disabled scenarios
func TestResult_IsEnabled(t *testing.T) {
	tests := []struct {
		name       string
		result     *Result
		wantEnabled bool
	}{
		{
			name: "variant key 'enabled'",
			result: &Result{
				VariantKey: "enabled",
			},
			wantEnabled: true,
		},
		{
			name: "variant key 'on'",
			result: &Result{
				VariantKey: "on",
			},
			wantEnabled: true,
		},
		{
			name: "variant key 'true'",
			result: &Result{
				VariantKey: "true",
			},
			wantEnabled: true,
		},
		{
			name: "variant key 'disabled'",
			result: &Result{
				VariantKey: "disabled",
			},
			wantEnabled: false,
		},
		{
			name: "attachment with enabled=true",
			result: &Result{
				VariantKey: "custom",
				VariantAttachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
				},
			},
			wantEnabled: true,
		},
		{
			name: "attachment with value=true",
			result: &Result{
				VariantKey: "custom",
				VariantAttachment: map[string]json.RawMessage{
					"value": json.RawMessage(`true`),
				},
			},
			wantEnabled: true,
		},
		{
			name: "attachment with enabled=false",
			result: &Result{
				VariantKey: "custom",
				VariantAttachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`false`),
				},
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.IsEnabled()
			assert.Equal(t, tt.wantEnabled, got)
		})
	}
}

// TestIsTrue tests various truthy/falsy values
func TestIsTrue(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
		want bool
	}{
		{
			name: "nil value",
			raw:  nil,
			want: false,
		},
		{
			name: "boolean true",
			raw:  json.RawMessage(`true`),
			want: true,
		},
		{
			name: "boolean false",
			raw:  json.RawMessage(`false`),
			want: false,
		},
		{
			name: "string 'true'",
			raw:  json.RawMessage(`"true"`),
			want: true,
		},
		{
			name: "string 'True' (case insensitive)",
			raw:  json.RawMessage(`"True"`),
			want: true,
		},
		{
			name: "string 'on'",
			raw:  json.RawMessage(`"on"`),
			want: true,
		},
		{
			name: "string 'enabled'",
			raw:  json.RawMessage(`"enabled"`),
			want: true,
		},
		{
			name: "string '1'",
			raw:  json.RawMessage(`"1"`),
			want: true,
		},
		{
			name: "string 'yes'",
			raw:  json.RawMessage(`"yes"`),
			want: true,
		},
		{
			name: "string 'false'",
			raw:  json.RawMessage(`"false"`),
			want: false,
		},
		{
			name: "string 'random'",
			raw:  json.RawMessage(`"random"`),
			want: false,
		},
		{
			name: "integer 1",
			raw:  json.RawMessage(`1`),
			want: true,
		},
		{
			name: "integer 0",
			raw:  json.RawMessage(`0`),
			want: false,
		},
		{
			name: "integer 2",
			raw:  json.RawMessage(`2`),
			want: false,
		},
		{
			name: "invalid json",
			raw:  json.RawMessage(`invalid`),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrue(tt.raw)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestResult_GetString tests string value extraction
func TestResult_GetString(t *testing.T) {
	tests := []struct {
		name       string
		result     *Result
		key        string
		defaultVal string
		want       string
	}{
		{
			name: "nil attachment",
			result: &Result{
				VariantAttachment: nil,
			},
			key:        "theme",
			defaultVal: "light",
			want:       "light",
		},
		{
			name: "key not found",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"color": json.RawMessage(`"blue"`),
				},
			},
			key:        "theme",
			defaultVal: "light",
			want:       "light",
		},
		{
			name: "valid string value",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"theme": json.RawMessage(`"dark"`),
				},
			},
			key:        "theme",
			defaultVal: "light",
			want:       "dark",
		},
		{
			name: "invalid json value",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"theme": json.RawMessage(`123`),
				},
			},
			key:        "theme",
			defaultVal: "light",
			want:       "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.GetString(tt.key, tt.defaultVal)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestResult_GetInt tests integer value extraction
func TestResult_GetInt(t *testing.T) {
	tests := []struct {
		name       string
		result     *Result
		key        string
		defaultVal int
		want       int
	}{
		{
			name: "nil attachment",
			result: &Result{
				VariantAttachment: nil,
			},
			key:        "limit",
			defaultVal: 100,
			want:       100,
		},
		{
			name: "key not found",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"other": json.RawMessage(`500`),
				},
			},
			key:        "limit",
			defaultVal: 100,
			want:       100,
		},
		{
			name: "valid integer value",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"limit": json.RawMessage(`1000`),
				},
			},
			key:        "limit",
			defaultVal: 100,
			want:       1000,
		},
		{
			name: "float64 value",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"limit": json.RawMessage(`500.5`),
				},
			},
			key:        "limit",
			defaultVal: 100,
			want:       500,
		},
		{
			name: "invalid json value",
			result: &Result{
				VariantAttachment: map[string]json.RawMessage{
					"limit": json.RawMessage(`"not a number"`),
				},
			},
			key:        "limit",
			defaultVal: 100,
			want:       100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.GetInt(tt.key, tt.defaultVal)
			assert.Equal(t, tt.want, got)
		})
	}
}
