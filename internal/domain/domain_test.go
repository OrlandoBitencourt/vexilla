package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// ----------------------
// EvaluationContext Tests
// ----------------------
//

func TestNewEvaluationContext(t *testing.T) {
	ctx := NewEvaluationContext("123")

	assert.Equal(t, "123", ctx.EntityID)
	assert.Equal(t, "user", ctx.EntityType)
	assert.NotNil(t, ctx.Context)
}

func TestEvaluationContext_Attributes(t *testing.T) {
	ctx := NewEvaluationContext("u1").
		WithAttribute("age", 20).
		WithEntityType("device")

	assert.Equal(t, "device", ctx.EntityType)

	val, ok := ctx.GetAttribute("age")
	assert.True(t, ok)
	assert.Equal(t, 20, val)

	assert.True(t, ctx.HasAttribute("age"))
	assert.False(t, ctx.HasAttribute("missing"))
}

func TestEvaluationContext_WithAttributes(t *testing.T) {
	ctx := NewEvaluationContext("x").WithAttributes(map[string]interface{}{
		"a": 1,
		"b": "ok",
	})

	v1, ok1 := ctx.GetAttribute("a")
	v2, ok2 := ctx.GetAttribute("b")

	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.Equal(t, 1, v1)
	assert.Equal(t, "ok", v2)
}

//
// ----------------------
// EvaluationResult Tests
// ----------------------
//

func TestEvaluationResult_IsEnabled(t *testing.T) {
	r1 := &EvaluationResult{VariantKey: "on"}
	assert.True(t, r1.IsEnabled())

	r2 := &EvaluationResult{VariantAttachment: map[string]json.RawMessage{
		"enabled": json.RawMessage(`true`),
	}}
	assert.True(t, r2.IsEnabled())

	r3 := &EvaluationResult{VariantAttachment: map[string]json.RawMessage{
		"enabled": json.RawMessage(`false`),
	}}
	assert.False(t, r3.IsEnabled())
}

func TestEvaluationResult_GetStringValue(t *testing.T) {
	r := &EvaluationResult{VariantAttachment: map[string]json.RawMessage{
		"value": json.RawMessage(`"hello"`),
	}}

	assert.Equal(t, "hello", r.GetStringValue("default"))
	assert.Equal(t, "fallback", (&EvaluationResult{}).GetStringValue("fallback"))
}

func TestEvaluationResult_GetIntValue(t *testing.T) {
	r := &EvaluationResult{VariantAttachment: map[string]json.RawMessage{
		"value": json.RawMessage(`42`),
	}}

	assert.Equal(t, 42, r.GetIntValue(0))
}

func TestEvaluationResult_GetFloatValue(t *testing.T) {
	r := &EvaluationResult{VariantAttachment: map[string]json.RawMessage{
		"value": json.RawMessage(`12.5`),
	}}

	assert.Equal(t, 12.5, r.GetFloatValue(0))
}

func TestEvaluationResult_GetJSONValue(t *testing.T) {
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"config": json.RawMessage(`{"x": 1}`),
		},
	}

	v := r.GetJSONValue("config")
	require.NotNil(t, v)

	m := v.(map[string]interface{})
	assert.Equal(t, 1.0, m["x"])
}

//
// ----------------------
// Flag Tests
// ----------------------
//

func TestFlag_Validate(t *testing.T) {
	flag := &Flag{
		Key: "test",
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{Percent: 100, VariantID: 1},
				},
			},
		},
	}

	require.NoError(t, flag.Validate())
}

func TestFlag_Validate_NoKey(t *testing.T) {
	flag := &Flag{}
	assert.Error(t, flag.Validate())
}

func TestSegment_Validate_InvalidPercent(t *testing.T) {
	seg := Segment{
		RolloutPercent: 150,
		Distributions: []Distribution{
			{Percent: 100, VariantID: 1},
		},
	}

	assert.Error(t, seg.Validate())
}

func TestSegment_Validate_DistributionSum(t *testing.T) {
	seg := Segment{
		RolloutPercent: 100,
		Distributions: []Distribution{
			{Percent: 50, VariantID: 1},
			{Percent: 40, VariantID: 2},
		},
	}

	assert.Error(t, seg.Validate())
}

func TestFlag_GetVariantByID(t *testing.T) {
	flag := Flag{
		Variants: []Variant{
			{ID: 1, Key: "v1"},
		},
	}

	v, ok := flag.GetVariantByID(1)
	assert.True(t, ok)
	assert.Equal(t, "v1", v.Key)

	_, ok = flag.GetVariantByID(999)
	assert.False(t, ok)
}

func TestFlag_GetDefaultValue(t *testing.T) {
	flag := Flag{
		Enabled: true,
		Segments: []Segment{
			{
				Distributions: []Distribution{
					{VariantID: 1, Percent: 100},
				},
			},
		},
		Variants: []Variant{
			{ID: 1, Attachment: map[string]json.RawMessage{"x": json.RawMessage(`1`)}},
		},
	}

	val := flag.GetDefaultValue()
	require.NotNil(t, val)

	m := val.(map[string]json.RawMessage)
	assert.Equal(t, `1`, string(m["x"]))
}

func TestFlag_SortedSegments(t *testing.T) {
	flag := Flag{
		Segments: []Segment{
			{Rank: 3},
			{Rank: 1},
			{Rank: 2},
		},
	}

	sorted := flag.SortedSegments()
	assert.Equal(t, 1, sorted[0].Rank)
	assert.Equal(t, 2, sorted[1].Rank)
	assert.Equal(t, 3, sorted[2].Rank)
}

//
// ----------------------
// Errors Tests
// ----------------------
//

func TestValidationError(t *testing.T) {
	err := NewValidationError("bad")
	assert.Equal(t, "validation error: bad", err.Error())
}

func TestCircuitOpenError(t *testing.T) {
	err := NewCircuitOpenError("open")
	assert.Equal(t, "circuit open: open", err.Error())
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("missing", "abc")
	assert.Equal(t, "missing not found: abc", err.Error())
}

func TestEvaluationContext_OverwriteAttribute(t *testing.T) {
	ctx := NewEvaluationContext("u").WithAttribute("x", 1)
	ctx = ctx.WithAttribute("x", 2)

	v, ok := ctx.GetAttribute("x")
	require.True(t, ok)
	assert.Equal(t, 2, v)
}

func TestEvaluationContext_GetAttribute_WhenNilMap(t *testing.T) {
	ctx := EvaluationContext{EntityID: "a", EntityType: "user", Context: nil}

	v, ok := ctx.GetAttribute("anything")
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestEvaluationContext_WithAttributesOnNil(t *testing.T) {
	ctx := EvaluationContext{EntityID: "x", Context: nil}
	ctx = ctx.WithAttributes(map[string]interface{}{"a": 123})

	v, ok := ctx.GetAttribute("a")
	assert.True(t, ok)
	assert.Equal(t, 123, v)
}

func TestEvaluationResult_IsEnabled_InvalidJSON(t *testing.T) {
	// JSON inválido -> Should NOT panic, should return false
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"enabled": json.RawMessage(`not-valid-json`),
		},
	}

	assert.False(t, r.IsEnabled())
}

func TestEvaluationResult_GetStringValue_InvalidJSON(t *testing.T) {
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"value": json.RawMessage(`{broken-json`),
		},
	}

	assert.Equal(t, "fallback", r.GetStringValue("fallback"))
}

func TestEvaluationResult_GetIntValue_Invalid(t *testing.T) {
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"value": json.RawMessage(`"not-a-number"`),
		},
	}

	assert.Equal(t, 7, r.GetIntValue(7))
}

func TestEvaluationResult_GetFloatValue_Invalid(t *testing.T) {
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"value": json.RawMessage(`"x"`),
		},
	}

	assert.Equal(t, 1.5, r.GetFloatValue(1.5))
}

func TestEvaluationResult_GetJSONValue_Invalid(t *testing.T) {
	r := &EvaluationResult{
		VariantAttachment: map[string]json.RawMessage{
			"config": json.RawMessage(`invalid{json`),
		},
	}

	v := r.GetJSONValue("config")
	assert.Nil(t, v)
}

func TestFlag_Validate_DisabledFlag(t *testing.T) {
	flag := &Flag{
		Key:     "disabled-flag",
		Enabled: false,
	}

	// Flags sem segments ainda são válidas
	assert.NoError(t, flag.Validate())
}

func TestFlag_Validate_DistributionSumTooLow(t *testing.T) {
	flag := &Flag{
		Key: "test",
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{Percent: 30},
					{Percent: 30},
				}, // Total = 60 → inválido
			},
		},
	}

	assert.Error(t, flag.Validate())
}

func TestFlag_Validate_InvalidDistributionPercent(t *testing.T) {
	flag := &Flag{
		Key: "x",
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{Percent: -1}, // inválido
				},
			},
		},
	}

	assert.Error(t, flag.Validate())
}

func TestFlag_Validate_DistributionVariantMissing(t *testing.T) {
	flag := &Flag{
		Key: "flag",
		Segments: []Segment{
			{
				RolloutPercent: 100,
				Distributions: []Distribution{
					{VariantID: 999, Percent: 100},
				},
			},
		},
		Variants: []Variant{
			{ID: 1, Key: "v1"},
		},
	}

	assert.Error(t, flag.Validate())
}

func TestSegment_Validate_NegativeRollout(t *testing.T) {
	seg := Segment{RolloutPercent: -10}
	assert.Error(t, seg.Validate())
}

func TestSegment_Validate_NegativeDistribution(t *testing.T) {
	seg := Segment{
		RolloutPercent: 100,
		Distributions: []Distribution{
			{Percent: -5},
		},
	}
	assert.Error(t, seg.Validate())
}

func TestSegment_Validate_EmptyDistribution(t *testing.T) {
	seg := Segment{
		RolloutPercent: 100,
		Distributions:  nil,
	}
	assert.Error(t, seg.Validate())
}

func TestFlag_GetDefaultValue_NoSegments(t *testing.T) {
	flag := Flag{
		Enabled:  true,
		Segments: nil,
	}

	assert.Nil(t, flag.GetDefaultValue())
}

func TestFlag_GetDefaultValue_DistributionVariantMissing(t *testing.T) {
	flag := Flag{
		Enabled: true,
		Segments: []Segment{
			{
				Distributions: []Distribution{{VariantID: 999}},
			},
		},
		Variants: []Variant{{ID: 1}},
	}

	assert.Nil(t, flag.GetDefaultValue())
}

func TestEvaluationError(t *testing.T) {
	err := NewEvaluationError("flag123", "failed hard", assert.AnError)

	msg := err.Error()
	assert.Contains(t, msg, "evaluation error")
	assert.Contains(t, msg, "flag123")
	assert.Contains(t, msg, "failed hard")
}
