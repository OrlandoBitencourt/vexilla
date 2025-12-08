package evaluator

import (
	"context"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_Evaluate(t *testing.T) {
	eval := NewEvaluator()
	ctx := context.Background()

	tests := []struct {
		name            string
		flag            vexilla.Flag
		evalCtx         vexilla.EvaluationContext
		expectedVariant string
		shouldMatch     bool
	}{
		{
			name: "disabled flag returns disabled",
			flag: vexilla.Flag{
				Key:     "disabled_flag",
				Enabled: false,
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
			},
			expectedVariant: "disabled",
			shouldMatch:     true,
		},
		{
			name: "simple EQ constraint - match",
			flag: vexilla.Flag{
				Key:     "country_flag",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "EQ", Value: "BR"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
				Context: map[string]interface{}{
					"country": "BR",
				},
			},
			expectedVariant: "enabled",
			shouldMatch:     true,
		},
		{
			name: "simple EQ constraint - no match",
			flag: vexilla.Flag{
				Key:     "country_flag",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "EQ", Value: "BR"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
				Context: map[string]interface{}{
					"country": "US",
				},
			},
			expectedVariant: "default",
			shouldMatch:     false,
		},
		{
			name: "IN operator - match",
			flag: vexilla.Flag{
				Key:     "multi_country",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "IN", Value: []interface{}{"US", "CA", "UK"}},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
				Context: map[string]interface{}{
					"country": "CA",
				},
			},
			expectedVariant: "enabled",
			shouldMatch:     true,
		},
		{
			name: "multiple constraints - all must match",
			flag: vexilla.Flag{
				Key:     "premium_us",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "EQ", Value: "US"},
							{Property: "tier", Operator: "EQ", Value: "premium"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
				Context: map[string]interface{}{
					"country": "US",
					"tier":    "premium",
				},
			},
			expectedVariant: "enabled",
			shouldMatch:     true,
		},
		{
			name: "regex match - document validation",
			flag: vexilla.Flag{
				Key:     "brazil_document",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						Constraints: []vexilla.Constraint{
							{Property: "document", Operator: "MATCHES", Value: "[0-9]{7}(0[0-9]|10)$"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			evalCtx: vexilla.EvaluationContext{
				EntityID: "user123",
				Context: map[string]interface{}{
					"document": "1234567808",
				},
			},
			expectedVariant: "enabled",
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(ctx, tt.flag, tt.evalCtx)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVariant, result.VariantKey)
			assert.True(t, result.EvaluatedLocally)
		})
	}
}

func TestEvaluator_BuildEnvironment(t *testing.T) {
	eval := NewEvaluator()

	evalCtx := vexilla.EvaluationContext{
		EntityID:   "user123",
		EntityType: "user",
		Context: map[string]interface{}{
			"country": "BR",
			"age":     25,
			"premium": true,
		},
	}

	env := eval.buildEnvironment(evalCtx)

	assert.Equal(t, "user123", env["entityID"])
	assert.Equal(t, "user", env["entityType"])
	assert.Equal(t, "BR", env["country"])
	assert.Equal(t, 25, env["age"])
	assert.Equal(t, true, env["premium"])
}
