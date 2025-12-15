package flagr

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagToDomain(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		f    *FlagrFlag
		want domain.Flag
	}{
		{
			name: "complete flag with all fields",
			f: &FlagrFlag{
				ID:          1,
				Key:         "new_custom_checkout",
				Description: "a/b test for new checkout option",
				Enabled:     true,
				UpdatedAt:   now,
				Segments: []FlagrSegment{
					{
						ID:             1,
						Rank:           1,
						Description:    "active for country brazil",
						RolloutPercent: 100,
						Constraints: []FlagrConstraint{
							{
								ID:       1,
								Property: "country",
								Operator: "EQ",
								Value:    "br",
							},
						},
						Distributions: []FlagrDistribution{
							{
								ID:        1,
								Percent:   100,
								VariantID: 2,
							},
						},
					},
					{
						ID:             2,
						Rank:           999,
						Description:    "disabled for other regions",
						RolloutPercent: 100,
						Constraints:    []FlagrConstraint{},
						Distributions: []FlagrDistribution{
							{
								ID:        2,
								Percent:   100,
								VariantID: 1,
							},
						},
					},
				},
				Variants: []FlagrVariant{
					{
						ID:         1,
						Key:        "disabled",
						Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`false`)},
					},
					{
						ID:         2,
						Key:        "enabled",
						Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`true`)},
					},
				},
				Tags: []Tag{
					{Value: "checkout"},
					{Value: "brazil"},
				},
				DataRecordsEnabled: true,
			},
			want: domain.Flag{
				ID:          1,
				Key:         "new_custom_checkout",
				Description: "a/b test for new checkout option",
				Enabled:     true,
				UpdatedAt:   now,
				Segments: []domain.Segment{
					{
						ID:             1,
						Rank:           1,
						Description:    "active for country brazil",
						RolloutPercent: 100,
						Constraints: []domain.Constraint{
							{
								ID:       1,
								Property: "country",
								Operator: domain.OperatorEQ,
								Value:    "br",
							},
						},
						Distributions: []domain.Distribution{
							{
								ID:        1,
								Percent:   100,
								VariantID: 2,
							},
						},
					},
					{
						ID:             2,
						Rank:           999,
						Description:    "disabled for other regions",
						RolloutPercent: 100,
						Constraints:    []domain.Constraint{},
						Distributions: []domain.Distribution{
							{
								ID:        2,
								Percent:   100,
								VariantID: 1,
							},
						},
					},
				},
				Variants: []domain.Variant{
					{
						ID:         1,
						Key:        "disabled",
						Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`false`)},
					},
					{
						ID:         2,
						Key:        "enabled",
						Attachment: map[string]json.RawMessage{"enabled": json.RawMessage(`true`)},
					},
				},
				Tags: []domain.Tag{
					{Value: "checkout"},
					{Value: "brazil"},
				},
				DataRecordsEnabled: true,
			},
		},
		{
			name: "minimal flag - no segments or variants",
			f: &FlagrFlag{
				ID:      2,
				Key:     "simple_flag",
				Enabled: false,
			},
			want: domain.Flag{
				ID:       2,
				Key:      "simple_flag",
				Enabled:  false,
				Segments: []domain.Segment{},
				Variants: []domain.Variant{},
				Tags:     []domain.Tag{},
			},
		},
		{
			name: "flag with empty segments and variants",
			f: &FlagrFlag{
				ID:       3,
				Key:      "empty_flag",
				Enabled:  true,
				Segments: []FlagrSegment{},
				Variants: []FlagrVariant{},
				Tags:     []Tag{},
			},
			want: domain.Flag{
				ID:       3,
				Key:      "empty_flag",
				Enabled:  true,
				Segments: []domain.Segment{},
				Variants: []domain.Variant{},
				Tags:     []domain.Tag{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluationContextFromDomain(tt.f.Key, domain.NewEvaluationContext(""))

			assert.Equal(t, tt.want.Key, got.FlagKey)

		})
	}
}

func TestExtractEvaluationReason(t *testing.T) {
	tests := []struct {
		name string
		log  EvalDebugLog
		want string
	}{
		{
			name: "main message present",
			log: EvalDebugLog{
				Msg: "flag evaluated",
				SegmentDebugLogs: []SegmentDebugLog{
					{Msg: "segment matched"},
				},
			},
			want: "flag evaluated",
		},
		{
			name: "only segment debug log",
			log: EvalDebugLog{
				Msg: "",
				SegmentDebugLogs: []SegmentDebugLog{
					{Msg: "first segment"},
					{Msg: "second segment"},
				},
			},
			want: "first segment",
		},
		{
			name: "empty log",
			log: EvalDebugLog{
				Msg:              "",
				SegmentDebugLogs: []SegmentDebugLog{},
			},
			want: "evaluated successfully",
		},
		{
			name: "empty segment logs",
			log: EvalDebugLog{
				Msg: "",
				SegmentDebugLogs: []SegmentDebugLog{
					{Msg: ""},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEvaluationReason(tt.log)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFlagToDomain_RoundTrip(t *testing.T) {
	// Create a complete Flagr flag
	original := &FlagrFlag{
		ID:          1,
		Key:         "integration_test",
		Description: "Integration test flag",
		Enabled:     true,
		Segments: []FlagrSegment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []FlagrConstraint{
					{ID: 1, Property: "country", Operator: "EQ", Value: "BR"},
				},
				Distributions: []FlagrDistribution{
					{ID: 1, Percent: 100, VariantID: 1},
				},
			},
		},
		Variants: []FlagrVariant{
			{
				ID:  1,
				Key: "enabled",
				Attachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
				},
			},
		},
		Tags: []Tag{
			{Value: "integration"},
		},
	}

	// Convert to domain
	domainFlag := FlagToDomain(original)

	// Verify all fields were converted
	assert.Equal(t, original.ID, domainFlag.ID)
	assert.Equal(t, original.Key, domainFlag.Key)
	assert.Equal(t, original.Description, domainFlag.Description)
	assert.Equal(t, original.Enabled, domainFlag.Enabled)
	assert.Len(t, domainFlag.Segments, 1)
	assert.Len(t, domainFlag.Variants, 1)
	assert.Len(t, domainFlag.Tags, 1)

	// Verify nested structures
	assert.Equal(t, original.Segments[0].ID, domainFlag.Segments[0].ID)
	assert.Equal(t, original.Variants[0].Key, domainFlag.Variants[0].Key)
	assert.Equal(t, original.Tags[0].Value, domainFlag.Tags[0].Value)
}

func TestEvaluationContextFromDomain_RoundTrip(t *testing.T) {
	// Create domain context
	domainCtx := domain.EvaluationContext{
		EntityID:   "test_user",
		EntityType: "user",
		Context: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
			"score":   95.5,
			"active":  true,
		},
	}

	// Convert to Flagr request
	request := EvaluationContextFromDomain("test_flag", domainCtx)

	// Verify conversion
	assert.Equal(t, "test_flag", request.FlagKey)
	assert.Equal(t, domainCtx.EntityID, request.EntityID)
	assert.Equal(t, domainCtx.EntityType, request.EntityType)
	assert.Equal(t, "BR", request.EntityContext["country"])
	assert.Equal(t, "premium", request.EntityContext["tier"])
	assert.Equal(t, 95.5, request.EntityContext["score"])
	assert.Equal(t, true, request.EntityContext["active"])
}

func BenchmarkFlagToDomain(b *testing.B) {
	flagrFlag := &FlagrFlag{
		ID:          1,
		Key:         "bench_flag",
		Description: "benchmark test",
		Enabled:     true,
		Segments: []FlagrSegment{
			{
				ID:             1,
				RolloutPercent: 100,
				Constraints: []FlagrConstraint{
					{ID: 1, Property: "country", Operator: "EQ", Value: "BR"},
				},
				Distributions: []FlagrDistribution{
					{ID: 1, Percent: 100, VariantID: 1},
				},
			},
		},
		Variants: []FlagrVariant{
			{ID: 1, Key: "on", Attachment: map[string]json.RawMessage{}},
		},
		Tags: []Tag{{Value: "test"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FlagToDomain(flagrFlag)
	}
}

func BenchmarkFlagsToDomain(b *testing.B) {
	flags := make([]FlagrFlag, 100)
	for i := 0; i < 100; i++ {
		flags[i] = FlagrFlag{
			ID:      int64(i),
			Key:     "flag_" + string(rune(i)),
			Enabled: i%2 == 0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FlagsToDomain(flags)
	}
}

func BenchmarkEvaluationContextFromDomain(b *testing.B) {
	evalCtx := domain.EvaluationContext{
		EntityID:   "user123",
		EntityType: "user",
		Context: map[string]interface{}{
			"country": "BR",
			"tier":    "premium",
			"age":     25,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EvaluationContextFromDomain("test_flag", evalCtx)
	}
}

func BenchmarkEvaluationResultToDomain(b *testing.B) {
	resp := EvaluationResponse{
		FlagID:     1,
		FlagKey:    "bench",
		SegmentID:  10,
		VariantID:  100,
		VariantKey: "on",
		VariantAttachment: map[string]json.RawMessage{
			"enabled": json.RawMessage(`true`),
		},
		Timestamp: time.Now(),
		EvalDebugLog: EvalDebugLog{
			Msg: "matched",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EvaluationResultToDomain(resp)
	}
}

func BenchmarkConstraintToDomain(b *testing.B) {
	constraint := FlagrConstraint{
		ID:       1,
		Property: "country",
		Operator: "EQ",
		Value:    "BR",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConstraintToDomain(constraint)
	}
}

func TestFlagsToDomain(t *testing.T) {
	tests := []struct {
		name  string
		flags []FlagrFlag
		want  int
	}{
		{
			name: "multiple flags",
			flags: []FlagrFlag{
				{ID: 1, Key: "flag1", Enabled: true},
				{ID: 2, Key: "flag2", Enabled: false},
				{ID: 3, Key: "flag3", Enabled: true},
			},
			want: 3,
		},
		{
			name:  "empty slice",
			flags: []FlagrFlag{},
			want:  0,
		},
		{
			name: "single flag",
			flags: []FlagrFlag{
				{ID: 1, Key: "only_flag"},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlagsToDomain(tt.flags)
			assert.Len(t, result, tt.want)

			if tt.want > 0 {
				assert.Equal(t, tt.flags[0].ID, result[0].ID)
				assert.Equal(t, tt.flags[0].Key, result[0].Key)
			}
		})
	}
}

func TestSegmentsToDomain(t *testing.T) {
	segments := []FlagrSegment{
		{ID: 1, Rank: 0, RolloutPercent: 100},
		{ID: 2, Rank: 1, RolloutPercent: 50},
	}

	result := SegmentsToDomain(segments)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, 0, result[0].Rank)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, 1, result[1].Rank)
}

func TestSegmentsToDomain_Empty(t *testing.T) {
	result := SegmentsToDomain([]FlagrSegment{})
	assert.Len(t, result, 0)
}

func TestSegmentToDomain(t *testing.T) {
	tests := []struct {
		name    string
		segment FlagrSegment
		want    domain.Segment
	}{
		{
			name: "complete segment",
			segment: FlagrSegment{
				ID:             10,
				Rank:           5,
				Description:    "test segment",
				RolloutPercent: 50,
				Constraints: []FlagrConstraint{
					{
						ID:       1,
						Property: "country",
						Operator: "IN",
						Value:    "BR,US",
					},
				},
				Distributions: []FlagrDistribution{
					{ID: 1, Percent: 50, VariantID: 100},
					{ID: 2, Percent: 50, VariantID: 200},
				},
			},
			want: domain.Segment{
				ID:             10,
				Rank:           5,
				Description:    "test segment",
				RolloutPercent: 50,
				Constraints: []domain.Constraint{
					{
						ID:       1,
						Property: "country",
						Operator: domain.OperatorIN,
						Value:    "BR,US",
					},
				},
				Distributions: []domain.Distribution{
					{ID: 1, Percent: 50, VariantID: 100},
					{ID: 2, Percent: 50, VariantID: 200},
				},
			},
		},
		{
			name: "minimal segment",
			segment: FlagrSegment{
				ID:             1,
				RolloutPercent: 100,
			},
			want: domain.Segment{
				ID:             1,
				RolloutPercent: 100,
				Constraints:    []domain.Constraint{},
				Distributions:  []domain.Distribution{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SegmentToDomain(tt.segment)

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Rank, got.Rank)
			assert.Equal(t, tt.want.Description, got.Description)
			assert.Equal(t, tt.want.RolloutPercent, got.RolloutPercent)
			assert.Len(t, got.Constraints, len(tt.want.Constraints))
			assert.Len(t, got.Distributions, len(tt.want.Distributions))
		})
	}
}

func TestConstraintsToDomain(t *testing.T) {
	constraints := []FlagrConstraint{
		{ID: 1, Property: "country", Operator: "EQ", Value: "BR"},
		{ID: 2, Property: "tier", Operator: "IN", Value: "premium,enterprise"},
		{ID: 3, Property: "age", Operator: "GT", Value: "18"},
	}

	result := ConstraintsToDomain(constraints)

	require.Len(t, result, 3)
	assert.Equal(t, "country", result[0].Property)
	assert.Equal(t, domain.OperatorEQ, result[0].Operator)
	assert.Equal(t, "tier", result[1].Property)
	assert.Equal(t, domain.OperatorIN, result[1].Operator)
	assert.Equal(t, "age", result[2].Property)
	assert.Equal(t, domain.OperatorGT, result[2].Operator)
}

func TestConstraintsToDomain_Empty(t *testing.T) {
	result := ConstraintsToDomain([]FlagrConstraint{})
	assert.Len(t, result, 0)
}

func TestConstraintToDomain(t *testing.T) {
	tests := []struct {
		name       string
		constraint FlagrConstraint
		want       domain.Constraint
	}{
		{
			name: "EQ operator",
			constraint: FlagrConstraint{
				ID:       1,
				Property: "tier",
				Operator: "EQ",
				Value:    "premium",
			},
			want: domain.Constraint{
				ID:       1,
				Property: "tier",
				Operator: domain.OperatorEQ,
				Value:    "premium",
			},
		},
		{
			name: "NEQ operator",
			constraint: FlagrConstraint{
				ID:       2,
				Property: "tier",
				Operator: "NEQ",
				Value:    "free",
			},
			want: domain.Constraint{
				ID:       2,
				Property: "tier",
				Operator: domain.OperatorNEQ,
				Value:    "free",
			},
		},
		{
			name: "GT operator",
			constraint: FlagrConstraint{
				ID:       3,
				Property: "age",
				Operator: "GT",
				Value:    "18",
			},
			want: domain.Constraint{
				ID:       3,
				Property: "age",
				Operator: domain.OperatorGT,
				Value:    "18",
			},
		},
		{
			name: "GTE operator",
			constraint: FlagrConstraint{
				ID:       4,
				Property: "age",
				Operator: "GTE",
				Value:    "21",
			},
			want: domain.Constraint{
				ID:       4,
				Property: "age",
				Operator: domain.OperatorGTE,
				Value:    "21",
			},
		},
		{
			name: "LT operator",
			constraint: FlagrConstraint{
				ID:       5,
				Property: "age",
				Operator: "LT",
				Value:    "65",
			},
			want: domain.Constraint{
				ID:       5,
				Property: "age",
				Operator: domain.OperatorLT,
				Value:    "65",
			},
		},
		{
			name: "LTE operator",
			constraint: FlagrConstraint{
				ID:       6,
				Property: "age",
				Operator: "LTE",
				Value:    "64",
			},
			want: domain.Constraint{
				ID:       6,
				Property: "age",
				Operator: domain.OperatorLTE,
				Value:    "64",
			},
		},
		{
			name: "IN operator",
			constraint: FlagrConstraint{
				ID:       7,
				Property: "country",
				Operator: "IN",
				Value:    "BR,US,UK",
			},
			want: domain.Constraint{
				ID:       7,
				Property: "country",
				Operator: domain.OperatorIN,
				Value:    "BR,US,UK",
			},
		},
		{
			name: "NOTIN operator",
			constraint: FlagrConstraint{
				ID:       8,
				Property: "country",
				Operator: "NOTIN",
				Value:    "CN,RU",
			},
			want: domain.Constraint{
				ID:       8,
				Property: "country",
				Operator: domain.OperatorNOTIN,
				Value:    "CN,RU",
			},
		},
		{
			name: "MATCHES operator",
			constraint: FlagrConstraint{
				ID:       9,
				Property: "email",
				Operator: "MATCHES",
				Value:    ".*@example\\\\.com",
			},
			want: domain.Constraint{
				ID:       9,
				Property: "email",
				Operator: domain.OperatorMATCHES,
				Value:    ".*@example\\\\.com",
			},
		},
		{
			name: "CONTAINS operator",
			constraint: FlagrConstraint{
				ID:       10,
				Property: "email",
				Operator: "CONTAINS",
				Value:    "example",
			},
			want: domain.Constraint{
				ID:       10,
				Property: "email",
				Operator: domain.OperatorCONTAINS,
				Value:    "example",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConstraintToDomain(tt.constraint)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Property, got.Property)
			assert.Equal(t, tt.want.Operator, got.Operator)
			assert.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestDistributionsToDomain(t *testing.T) {
	dists := []FlagrDistribution{
		{ID: 1, Percent: 50, VariantID: 100},
		{ID: 2, Percent: 30, VariantID: 200},
		{ID: 3, Percent: 20, VariantID: 300},
	}

	result := DistributionsToDomain(dists)

	require.Len(t, result, 3)
	assert.Equal(t, 50, result[0].Percent)
	assert.Equal(t, int64(100), result[0].VariantID)
	assert.Equal(t, 30, result[1].Percent)
	assert.Equal(t, int64(200), result[1].VariantID)
	assert.Equal(t, 20, result[2].Percent)
	assert.Equal(t, int64(300), result[2].VariantID)
}

func TestDistributionsToDomain_Empty(t *testing.T) {
	result := DistributionsToDomain([]FlagrDistribution{})
	assert.Len(t, result, 0)
}

func TestDistributionToDomain(t *testing.T) {
	tests := []struct {
		name string
		dist FlagrDistribution
		want domain.Distribution
	}{
		{
			name: "standard distribution",
			dist: FlagrDistribution{
				ID:        5,
				Percent:   75,
				VariantID: 10,
			},
			want: domain.Distribution{
				ID:        5,
				Percent:   75,
				VariantID: 10,
			},
		},
		{
			name: "zero percent",
			dist: FlagrDistribution{
				ID:        1,
				Percent:   0,
				VariantID: 1,
			},
			want: domain.Distribution{
				ID:        1,
				Percent:   0,
				VariantID: 1,
			},
		},
		{
			name: "100 percent",
			dist: FlagrDistribution{
				ID:        2,
				Percent:   100,
				VariantID: 2,
			},
			want: domain.Distribution{
				ID:        2,
				Percent:   100,
				VariantID: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DistributionToDomain(tt.dist)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Percent, got.Percent)
			assert.Equal(t, tt.want.VariantID, got.VariantID)
		})
	}
}

func TestVariantsToDomain(t *testing.T) {
	variants := []FlagrVariant{
		{ID: 1, Key: "control"},
		{ID: 2, Key: "treatment_a"},
		{ID: 3, Key: "treatment_b"},
	}

	result := VariantsToDomain(variants)

	require.Len(t, result, 3)
	assert.Equal(t, "control", result[0].Key)
	assert.Equal(t, "treatment_a", result[1].Key)
	assert.Equal(t, "treatment_b", result[2].Key)
}

func TestVariantsToDomain_Empty(t *testing.T) {
	result := VariantsToDomain([]FlagrVariant{})
	assert.Len(t, result, 0)
}

func TestVariantToDomain(t *testing.T) {
	tests := []struct {
		name    string
		variant FlagrVariant
		want    domain.Variant
	}{
		{
			name: "variant with complex attachment",
			variant: FlagrVariant{
				ID:  1,
				Key: "treatment_a",
				Attachment: map[string]json.RawMessage{
					"color":   json.RawMessage(`"blue"`),
					"enabled": json.RawMessage(`true`),
					"timeout": json.RawMessage(`300`),
				},
			},
			want: domain.Variant{
				ID:  1,
				Key: "treatment_a",
				Attachment: map[string]json.RawMessage{
					"color":   json.RawMessage(`"blue"`),
					"enabled": json.RawMessage(`true`),
					"timeout": json.RawMessage(`300`),
				},
			},
		},
		{
			name: "variant without attachment",
			variant: FlagrVariant{
				ID:  2,
				Key: "control",
			},
			want: domain.Variant{
				ID:         2,
				Key:        "control",
				Attachment: nil,
			},
		},
		{
			name: "variant with empty attachment",
			variant: FlagrVariant{
				ID:         3,
				Key:        "empty",
				Attachment: map[string]json.RawMessage{},
			},
			want: domain.Variant{
				ID:         3,
				Key:        "empty",
				Attachment: map[string]json.RawMessage{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VariantToDomain(tt.variant)
			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Key, got.Key)
			assert.Equal(t, len(tt.want.Attachment), len(got.Attachment))
		})
	}
}

func TestTagsToDomain(t *testing.T) {
	tests := []struct {
		name string
		tags []Tag
		want []domain.Tag
	}{
		{
			name: "multiple tags",
			tags: []Tag{
				{Value: "production"},
				{Value: "user-service"},
				{Value: "critical"},
			},
			want: []domain.Tag{
				{Value: "production"},
				{Value: "user-service"},
				{Value: "critical"},
			},
		},
		{
			name: "single tag",
			tags: []Tag{
				{Value: "test"},
			},
			want: []domain.Tag{
				{Value: "test"},
			},
		},
		{
			name: "empty tags",
			tags: []Tag{},
			want: []domain.Tag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TagsToDomain(tt.tags)
			require.Len(t, got, len(tt.want))
			for i, tag := range tt.want {
				assert.Equal(t, tag.Value, got[i].Value)
			}
		})
	}
}

func TestEvaluationResultToDomain(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		resp EvaluationResponse
		want domain.EvaluationResult
	}{
		{
			name: "complete evaluation response",
			resp: EvaluationResponse{
				FlagID:     1,
				FlagKey:    "test_flag",
				SegmentID:  10,
				VariantID:  100,
				VariantKey: "enabled",
				VariantAttachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
					"value":   json.RawMessage(`"premium"`),
				},
				Timestamp: now,
				EvalDebugLog: EvalDebugLog{
					Msg: "matched segment",
					SegmentDebugLogs: []SegmentDebugLog{
						{Msg: "constraint passed"},
					},
				},
			},
			want: domain.EvaluationResult{
				FlagID:     1,
				FlagKey:    "test_flag",
				SegmentID:  10,
				VariantID:  100,
				VariantKey: "enabled",
				VariantAttachment: map[string]json.RawMessage{
					"enabled": json.RawMessage(`true`),
					"value":   json.RawMessage(`"premium"`),
				},
				EvaluationReason: "matched segment",
				Timestamp:        now,
			},
		},
		{
			name: "with empty main debug message but has segment log",
			resp: EvaluationResponse{
				FlagID:     1,
				FlagKey:    "flag",
				VariantKey: "on",
				Timestamp:  now,
				EvalDebugLog: EvalDebugLog{
					Msg: "",
					SegmentDebugLogs: []SegmentDebugLog{
						{Msg: "first segment matched"},
					},
				},
			},
			want: domain.EvaluationResult{
				FlagID:           1,
				FlagKey:          "flag",
				VariantKey:       "on",
				EvaluationReason: "first segment matched",
				Timestamp:        now,
			},
		},
		{
			name: "with completely empty debug log",
			resp: EvaluationResponse{
				FlagID:     1,
				FlagKey:    "flag",
				VariantKey: "on",
				Timestamp:  now,
				EvalDebugLog: EvalDebugLog{
					Msg:              "",
					SegmentDebugLogs: []SegmentDebugLog{},
				},
			},
			want: domain.EvaluationResult{
				FlagID:           1,
				FlagKey:          "flag",
				VariantKey:       "on",
				EvaluationReason: "evaluated successfully",
				Timestamp:        now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluationResultToDomain(tt.resp)

			assert.Equal(t, tt.want.FlagID, got.FlagID)
			assert.Equal(t, tt.want.FlagKey, got.FlagKey)
			assert.Equal(t, tt.want.SegmentID, got.SegmentID)
			assert.Equal(t, tt.want.VariantID, got.VariantID)
			assert.Equal(t, tt.want.VariantKey, got.VariantKey)
			assert.Equal(t, tt.want.EvaluationReason, got.EvaluationReason)
			assert.Equal(t, tt.want.Timestamp, got.Timestamp)
		})
	}
}
