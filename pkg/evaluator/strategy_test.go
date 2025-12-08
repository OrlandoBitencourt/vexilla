package evaluator

import (
	"testing"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/stretchr/testify/assert"
)

func TestDeterminer_Determine(t *testing.T) {
	determiner := NewDeterminer()

	tests := []struct {
		name     string
		flag     vexilla.Flag
		expected Strategy
		reason   string
	}{
		{
			name: "disabled flag - local",
			flag: vexilla.Flag{
				Key:     "disabled",
				Enabled: false,
			},
			expected: StrategyLocal,
			reason:   "disabled flags always evaluate to default locally",
		},
		{
			name: "no segments - local",
			flag: vexilla.Flag{
				Key:      "no_segments",
				Enabled:  true,
				Segments: []vexilla.Segment{},
			},
			expected: StrategyLocal,
			reason:   "no segments means always default",
		},
		{
			name: "100% rollout with 100% distribution - local",
			flag: vexilla.Flag{
				Key:     "brazil_launch",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						RolloutPercent: 100,
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "EQ", Value: "BR"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			expected: StrategyLocal,
			reason:   "100% rollout with single 100% distribution is deterministic",
		},
		{
			name: "50% rollout - remote",
			flag: vexilla.Flag{
				Key:     "gradual_rollout",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						RolloutPercent: 50,
						Constraints: []vexilla.Constraint{
							{Property: "country", Operator: "EQ", Value: "US"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 100},
						},
					},
				},
			},
			expected: StrategyRemote,
			reason:   "partial rollout requires Flagr for consistent bucketing",
		},
		{
			name: "A/B test (multiple distributions) - remote",
			flag: vexilla.Flag{
				Key:     "ab_test",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						RolloutPercent: 100,
						Distributions: []vexilla.Distribution{
							{VariantKey: "control", Percent: 50},
							{VariantKey: "variant_a", Percent: 50},
						},
					},
				},
			},
			expected: StrategyRemote,
			reason:   "multiple distributions require Flagr for sticky assignment",
		},
		{
			name: "single distribution with 10% - remote",
			flag: vexilla.Flag{
				Key:     "beta_program",
				Enabled: true,
				Segments: []vexilla.Segment{
					{
						RolloutPercent: 100,
						Constraints: []vexilla.Constraint{
							{Property: "tier", Operator: "EQ", Value: "premium"},
						},
						Distributions: []vexilla.Distribution{
							{VariantKey: "enabled", Percent: 10},
						},
					},
				},
			},
			expected: StrategyRemote,
			reason:   "partial distribution percentage requires Flagr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determiner.Determine(tt.flag)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestDeterminer_CanEvaluateLocally(t *testing.T) {
	determiner := NewDeterminer()

	localFlag := vexilla.Flag{
		Key:     "local_flag",
		Enabled: true,
		Segments: []vexilla.Segment{
			{
				RolloutPercent: 100,
				Distributions: []vexilla.Distribution{
					{VariantKey: "on", Percent: 100},
				},
			},
		},
	}

	remoteFlag := vexilla.Flag{
		Key:     "remote_flag",
		Enabled: true,
		Segments: []vexilla.Segment{
			{
				RolloutPercent: 50,
				Distributions: []vexilla.Distribution{
					{VariantKey: "on", Percent: 100},
				},
			},
		},
	}

	assert.True(t, determiner.CanEvaluateLocally(localFlag))
	assert.False(t, determiner.CanEvaluateLocally(remoteFlag))
}
