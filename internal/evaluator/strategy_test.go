package evaluator

import (
	"testing"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewStrategyDeterminer(t *testing.T) {
	sd := NewStrategyDeterminer()
	assert.NotNil(t, sd)
}

func TestStrategyDeterminer_Determine(t *testing.T) {
	sd := NewStrategyDeterminer()

	tests := []struct {
		name     string
		flag     domain.Flag
		expected domain.EvaluationStrategy
	}{
		{
			name: "disabled flag - local",
			flag: domain.Flag{
				Enabled: false,
			},
			expected: domain.StrategyLocal,
		},
		{
			name: "no segments - local",
			flag: domain.Flag{
				Enabled:  true,
				Segments: []domain.Segment{},
			},
			expected: domain.StrategyLocal,
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
			expected: domain.StrategyLocal,
		},
		{
			name: "partial rollout - remote",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{
						RolloutPercent: 50,
						Distributions:  []domain.Distribution{{Percent: 100}},
					},
				},
			},
			expected: domain.StrategyRemote,
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
							{Percent: 50},
						},
					},
				},
			},
			expected: domain.StrategyRemote,
		},
		{
			name: "single distribution <100% - remote",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{
						RolloutPercent: 100,
						Distributions: []domain.Distribution{
							{Percent: 75},
						},
					},
				},
			},
			expected: domain.StrategyRemote,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sd.Determine(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStrategyDeterminer_RequiresFlagr(t *testing.T) {
	sd := NewStrategyDeterminer()

	tests := []struct {
		name     string
		segment  domain.Segment
		expected bool
	}{
		{
			name: "partial rollout",
			segment: domain.Segment{
				RolloutPercent: 50,
			},
			expected: true,
		},
		{
			name: "multiple distributions",
			segment: domain.Segment{
				RolloutPercent: 100,
				Distributions: []domain.Distribution{
					{Percent: 50},
					{Percent: 50},
				},
			},
			expected: true,
		},
		{
			name: "single distribution <100%",
			segment: domain.Segment{
				RolloutPercent: 100,
				Distributions: []domain.Distribution{
					{Percent: 80},
				},
			},
			expected: true,
		},
		{
			name: "100% deterministic",
			segment: domain.Segment{
				RolloutPercent: 100,
				Distributions: []domain.Distribution{
					{Percent: 100},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sd.requiresFlagr(tt.segment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStrategyDeterminer_IsLocalEvaluable(t *testing.T) {
	sd := NewStrategyDeterminer()

	flag := domain.Flag{
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 100}}},
		},
	}

	assert.True(t, sd.IsLocalEvaluable(flag))

	flag.Segments[0].RolloutPercent = 50
	assert.False(t, sd.IsLocalEvaluable(flag))
}

func TestStrategyDeterminer_IsRemoteEvaluable(t *testing.T) {
	sd := NewStrategyDeterminer()

	flag := domain.Flag{
		Enabled: true,
		Segments: []domain.Segment{
			{RolloutPercent: 50, Distributions: []domain.Distribution{{Percent: 100}}},
		},
	}

	assert.True(t, sd.IsRemoteEvaluable(flag))

	flag.Segments[0].RolloutPercent = 100
	assert.False(t, sd.IsRemoteEvaluable(flag))
}

func TestStrategyDeterminer_GetStrategyReason(t *testing.T) {
	sd := NewStrategyDeterminer()

	tests := []struct {
		name     string
		flag     domain.Flag
		contains string
	}{
		{
			name:     "disabled",
			flag:     domain.Flag{Enabled: false},
			contains: "disabled",
		},
		{
			name:     "no segments",
			flag:     domain.Flag{Enabled: true, Segments: []domain.Segment{}},
			contains: "no segments",
		},
		{
			name: "partial rollout",
			flag: domain.Flag{
				Enabled:  true,
				Segments: []domain.Segment{{RolloutPercent: 50}},
			},
			contains: "consistent hashing",
		},
		{
			name: "A/B test",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 50}, {Percent: 50}}},
				},
			},
			contains: "multiple variants",
		},
		{
			name: "percentage distribution",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 75}}},
				},
			},
			contains: "percentage-based",
		},
		{
			name: "100% deterministic",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 100}}},
				},
			},
			contains: "deterministic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := sd.GetStrategyReason(tt.flag)
			assert.Contains(t, reason, tt.contains)
		})
	}
}

func TestStrategyDeterminer_AnalyzeFlag(t *testing.T) {
	sd := NewStrategyDeterminer()

	flag := domain.Flag{
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				RolloutPercent: 50,
				Constraints: []domain.Constraint{
					{Property: "country", Operator: domain.OperatorEQ, Value: "US"},
				},
				Distributions: []domain.Distribution{
					{Percent: 50},
					{Percent: 50},
				},
			},
			{
				ID:             2,
				RolloutPercent: 100,
				Distributions: []domain.Distribution{
					{Percent: 100},
				},
			},
		},
	}

	analysis := sd.AnalyzeFlag(flag)

	assert.Equal(t, "test-flag", analysis.FlagKey)
	assert.True(t, analysis.Enabled)
	assert.Equal(t, domain.StrategyRemote, analysis.Strategy)
	assert.Len(t, analysis.Segments, 2)

	// Check first segment analysis
	seg1 := analysis.Segments[0]
	assert.Equal(t, int64(1), seg1.SegmentID)
	assert.Equal(t, 50, seg1.RolloutPercent)
	assert.Equal(t, 1, seg1.ConstraintCount)
	assert.Equal(t, 2, seg1.DistributionCount)
	assert.True(t, seg1.IsABTest)
	assert.True(t, seg1.RequiresFlagr)

	// Check second segment analysis
	seg2 := analysis.Segments[1]
	assert.Equal(t, int64(2), seg2.SegmentID)
	assert.Equal(t, 100, seg2.RolloutPercent)
	assert.Equal(t, 0, seg2.ConstraintCount)
	assert.Equal(t, 1, seg2.DistributionCount)
	assert.False(t, seg2.IsABTest)
	assert.False(t, seg2.RequiresFlagr)
}

func TestStrategyDeterminer_EstimatePerformance(t *testing.T) {
	sd := NewStrategyDeterminer()

	tests := []struct {
		name                string
		flag                domain.Flag
		expectedStrategy    domain.EvaluationStrategy
		expectedLatency     string
		expectedHTTPReqs    int
		expectedCacheNeeded bool
	}{
		{
			name: "local evaluation",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 100}}},
				},
			},
			expectedStrategy:    domain.StrategyLocal,
			expectedLatency:     "< 1ms",
			expectedHTTPReqs:    0,
			expectedCacheNeeded: true,
		},
		{
			name: "remote evaluation",
			flag: domain.Flag{
				Enabled: true,
				Segments: []domain.Segment{
					{RolloutPercent: 50},
				},
			},
			expectedStrategy:    domain.StrategyRemote,
			expectedLatency:     "50-200ms",
			expectedHTTPReqs:    1,
			expectedCacheNeeded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate := sd.EstimatePerformance(tt.flag)

			assert.Equal(t, tt.expectedStrategy, estimate.Strategy)
			assert.Equal(t, tt.expectedLatency, estimate.ExpectedLatency)
			assert.Equal(t, tt.expectedHTTPReqs, estimate.HTTPRequests)
			assert.Equal(t, tt.expectedCacheNeeded, estimate.CacheHitRequired)
			assert.Greater(t, estimate.ComplexityScore, 0)
		})
	}
}

func TestStrategyDeterminer_CalculateComplexity(t *testing.T) {
	sd := NewStrategyDeterminer()

	tests := []struct {
		name          string
		flag          domain.Flag
		minComplexity int
	}{
		{
			name: "simple flag",
			flag: domain.Flag{
				Segments: []domain.Segment{
					{Distributions: []domain.Distribution{{}}},
				},
			},
			minComplexity: 10,
		},
		{
			name: "complex flag with constraints",
			flag: domain.Flag{
				Segments: []domain.Segment{
					{
						Constraints: []domain.Constraint{
							{Property: "a"},
							{Property: "b"},
							{Property: "c"},
						},
						Distributions: []domain.Distribution{{}, {}},
					},
				},
			},
			minComplexity: 20,
		},
		{
			name: "flag with partial rollout",
			flag: domain.Flag{
				Segments: []domain.Segment{
					{
						RolloutPercent: 50,
						Distributions:  []domain.Distribution{{}},
					},
				},
			},
			minComplexity: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := sd.calculateComplexity(tt.flag)
			assert.GreaterOrEqual(t, complexity, tt.minComplexity)
		})
	}
}

func BenchmarkStrategyDeterminer_Determine(b *testing.B) {
	sd := NewStrategyDeterminer()

	flag := domain.Flag{
		Enabled: true,
		Segments: []domain.Segment{
			{
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{Property: "country", Operator: domain.OperatorEQ, Value: "US"},
					{Property: "tier", Operator: domain.OperatorIN, Value: "premium,enterprise"},
				},
				Distributions: []domain.Distribution{{Percent: 100}},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.Determine(flag)
	}
}

func BenchmarkStrategyDeterminer_AnalyzeFlag(b *testing.B) {
	sd := NewStrategyDeterminer()

	flag := domain.Flag{
		Key:     "bench-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{ID: 1, RolloutPercent: 50, Distributions: []domain.Distribution{{Percent: 50}, {Percent: 50}}},
			{ID: 2, RolloutPercent: 100, Distributions: []domain.Distribution{{Percent: 100}}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.AnalyzeFlag(flag)
	}
}
