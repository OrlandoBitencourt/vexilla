package evaluator

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create RawMessage
func raw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestNew(t *testing.T) {
	eval := New()
	assert.NotNil(t, eval)
	assert.NotNil(t, eval.programCache)
}

func TestEvaluator_Evaluate_DisabledFlag(t *testing.T) {
	eval := New()

	flag := domain.Flag{
		ID:      1,
		Key:     "disabled-flag",
		Enabled: false,
	}

	ctx := context.Background()
	result, err := eval.Evaluate(ctx, flag, domain.EvaluationContext{})

	require.NoError(t, err)
	assert.Equal(t, "flag disabled", result.EvaluationReason)
}

func TestEvaluator_Evaluate_NoSegments(t *testing.T) {
	eval := New()

	flag := domain.Flag{
		ID:       1,
		Key:      "no-segments",
		Enabled:  true,
		Segments: []domain.Segment{},
		Variants: []domain.Variant{
			{ID: 1, Key: "default", Attachment: map[string]json.RawMessage{"value": raw(true)}},
		},
	}

	ctx := context.Background()
	result, err := eval.Evaluate(ctx, flag, domain.EvaluationContext{})

	require.NoError(t, err)
	assert.Equal(t, "no segments", result.EvaluationReason)
	assert.Equal(t, int64(1), result.VariantID)
}

func TestEvaluator_Evaluate_SimpleMatch(t *testing.T) {
	eval := New()

	flag := domain.Flag{
		ID:      1,
		Key:     "country-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "country",
						Operator: domain.OperatorEQ,
						Value:    "BR",
					},
				},
				Distributions: []domain.Distribution{
					{ID: 1, VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "enabled", Attachment: map[string]json.RawMessage{"enabled": raw(true)}},
		},
	}

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "BR",
		},
	}

	result, err := eval.Evaluate(ctx, flag, evalCtx)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.SegmentID)
	assert.Equal(t, int64(1), result.VariantID)
	assert.Equal(t, "enabled", result.VariantKey)
	assert.Equal(t, raw(true), result.VariantAttachment["enabled"])
}

func TestEvaluator_Evaluate_NoMatch(t *testing.T) {
	eval := New()

	flag := domain.Flag{
		ID:      1,
		Key:     "country-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:   1,
				Rank: 0,
				Constraints: []domain.Constraint{
					{
						Property: "country",
						Operator: domain.OperatorEQ,
						Value:    "US",
					},
				},
				Distributions: []domain.Distribution{
					{ID: 1, VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "enabled", Attachment: map[string]json.RawMessage{"enabled": raw(true)}},
		},
	}

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "BR", // Doesn't match US
		},
	}

	result, err := eval.Evaluate(ctx, flag, evalCtx)

	require.NoError(t, err)
	assert.Equal(t, "no segments matched", result.EvaluationReason)
}

func TestEvaluator_EvaluateConstraint_Operators(t *testing.T) {
	eval := New()

	tests := []struct {
		name       string
		constraint domain.Constraint
		context    map[string]interface{}
		expected   bool
	}{
		{
			name: "EQ matches",
			constraint: domain.Constraint{
				Property: "tier",
				Operator: domain.OperatorEQ,
				Value:    "premium",
			},
			context:  map[string]interface{}{"tier": "premium"},
			expected: true,
		},
		{
			name: "EQ doesn't match",
			constraint: domain.Constraint{
				Property: "tier",
				Operator: domain.OperatorEQ,
				Value:    "premium",
			},
			context:  map[string]interface{}{"tier": "free"},
			expected: false,
		},
		{
			name: "NEQ matches",
			constraint: domain.Constraint{
				Property: "tier",
				Operator: domain.OperatorNEQ,
				Value:    "free",
			},
			context:  map[string]interface{}{"tier": "premium"},
			expected: true,
		},
		{
			name: "IN matches",
			constraint: domain.Constraint{
				Property: "country",
				Operator: domain.OperatorIN,
				Value:    []interface{}{"US", "BR", "UK"},
			},
			context:  map[string]interface{}{"country": "BR"},
			expected: true,
		},
		{
			name: "IN doesn't match",
			constraint: domain.Constraint{
				Property: "country",
				Operator: domain.OperatorIN,
				Value:    []interface{}{"US", "UK"},
			},
			context:  map[string]interface{}{"country": "BR"},
			expected: false,
		},
		{
			name: "NOTIN matches",
			constraint: domain.Constraint{
				Property: "country",
				Operator: domain.OperatorNOTIN,
				Value:    []interface{}{"US", "UK"},
			},
			context:  map[string]interface{}{"country": "BR"},
			expected: true,
		},
		{
			name: "GT matches",
			constraint: domain.Constraint{
				Property: "age",
				Operator: domain.OperatorGT,
				Value:    18,
			},
			context:  map[string]interface{}{"age": 25},
			expected: true,
		},
		{
			name: "LT matches",
			constraint: domain.Constraint{
				Property: "age",
				Operator: domain.OperatorLT,
				Value:    65,
			},
			context:  map[string]interface{}{"age": 25},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalCtx := domain.EvaluationContext{
				EntityID: "test",
				Context:  tt.context,
			}

			result, err := eval.evaluateConstraint(tt.constraint, evalCtx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_EvaluateConstraint_Regex(t *testing.T) {
	eval := New()

	constraint := domain.Constraint{
		Property: "email",
		Operator: domain.OperatorMATCHES,
		Value:    ".*@example\\\\.com$",
	}

	tests := []struct {
		email    string
		expected bool
	}{
		{"user@example.com", true},
		{"test@example.com", true},
		{"user@other.com", false},
		{"invalid-email", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			evalCtx := domain.EvaluationContext{
				EntityID: "test",
				Context:  map[string]interface{}{"email": tt.email},
			}

			result, err := eval.evaluateConstraint(constraint, evalCtx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result, "email: %s", tt.email)
		})
	}
}

func TestEvaluator_EvaluateSegment_MultipleConstraints(t *testing.T) {
	eval := New()

	segment := domain.Segment{
		ID: 1,
		Constraints: []domain.Constraint{
			{Property: "country", Operator: domain.OperatorEQ, Value: "BR"},
			{Property: "tier", Operator: domain.OperatorEQ, Value: "premium"},
			{Property: "age", Operator: domain.OperatorGT, Value: 18},
		},
	}

	tests := []struct {
		name     string
		context  map[string]interface{}
		expected bool
	}{
		{
			name: "all match",
			context: map[string]interface{}{
				"country": "BR",
				"tier":    "premium",
				"age":     25,
			},
			expected: true,
		},
		{
			name: "one doesn't match",
			context: map[string]interface{}{
				"country": "BR",
				"tier":    "free", // Doesn't match
				"age":     25,
			},
			expected: false,
		},
		{
			name: "missing property",
			context: map[string]interface{}{
				"country": "BR",
				// Missing tier and age
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evalCtx := domain.EvaluationContext{
				EntityID: "test",
				Context:  tt.context,
			}

			result, err := eval.evaluateSegment(segment, evalCtx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluator_CanEvaluateLocally(t *testing.T) {
	eval := New()

	tests := []struct {
		name     string
		flag     domain.Flag
		expected bool
	}{
		{
			name: "disabled flag - local",
			flag: domain.Flag{
				Enabled: false,
			},
			expected: true,
		},
		{
			name: "100% rollout - local",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{
						RolloutPercent: 100,
						Distributions:  []domain.Distribution{{Percent: 100}},
					},
				},
			},
			expected: true,
		},
		{
			name: "partial rollout - remote",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{
						RolloutPercent: 50, // Requires Flagr
						Distributions:  []domain.Distribution{{Percent: 100}},
					},
				},
			},
			expected: false,
		},
		{
			name: "A/B test - remote",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{
						RolloutPercent: 100,
						Distributions: []domain.Distribution{
							{Percent: 50},
							{Percent: 50}, // Multiple distributions
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.CanEvaluateLocally(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkEvaluator_Evaluate(b *testing.B) {
	eval := New()

	flag := domain.Flag{
		ID:      1,
		Key:     "bench-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:   1,
				Rank: 0,
				Constraints: []domain.Constraint{
					{Property: "country", Operator: domain.OperatorEQ, Value: "BR"},
					{Property: "tier", Operator: domain.OperatorEQ, Value: "premium"},
				},
				Distributions: []domain.Distribution{
					{ID: 1, VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "enabled", Attachment: map[string]json.RawMessage{"enabled": raw(true)}},
		},
	}

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user123",
		Context: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eval.Evaluate(ctx, flag, evalCtx)
	}
}
