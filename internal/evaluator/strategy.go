package evaluator

import (
	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// StrategyDeterminer determines the evaluation strategy for a flag
type StrategyDeterminer struct{}

// NewStrategyDeterminer creates a new strategy determiner
func NewStrategyDeterminer() *StrategyDeterminer {
	return &StrategyDeterminer{}
}

// Determine determines if a flag can be evaluated locally
// Returns true if the flag can be safely evaluated without Flagr's consistent hashing
func (s *StrategyDeterminer) Determine(flag domain.Flag) domain.EvaluationStrategy {
	// Rule 1: Disabled flags are always local
	if !flag.Enabled {
		return domain.StrategyLocal
	}

	// Rule 2: No segments = simple default value = local
	if len(flag.Segments) == 0 {
		return domain.StrategyLocal
	}

	// Rule 3: Check each segment for conditions requiring Flagr
	for _, segment := range flag.Segments {
		if s.requiresFlagr(segment) {
			return domain.StrategyRemote
		}
	}

	// All segments are 100% deterministic - can evaluate locally
	return domain.StrategyLocal
}

// requiresFlagr checks if a segment requires Flagr for evaluation
func (s *StrategyDeterminer) requiresFlagr(segment domain.Segment) bool {
	// Condition 1: Partial rollout (< 100%)
	// Requires Flagr's consistent hashing for sticky behavior
	if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
		return true
	}

	// Condition 2: Multiple distributions (A/B test)
	// Requires Flagr to assign users to variants consistently
	if len(segment.Distributions) > 1 {
		return true
	}

	// Condition 3: Single distribution with < 100% percentage
	// Requires Flagr for percentage-based allocation
	if len(segment.Distributions) == 1 {
		dist := segment.Distributions[0]
		if dist.Percent < 100 {
			return true
		}
	}

	// Segment is 100% deterministic based on constraints only
	return false
}

// IsLocalEvaluable is a convenience method that returns true if flag can be evaluated locally
func (s *StrategyDeterminer) IsLocalEvaluable(flag domain.Flag) bool {
	return s.Determine(flag) == domain.StrategyLocal
}

// IsRemoteEvaluable is a convenience method that returns true if flag requires remote evaluation
func (s *StrategyDeterminer) IsRemoteEvaluable(flag domain.Flag) bool {
	return s.Determine(flag) == domain.StrategyRemote
}

// GetStrategyReason returns a human-readable explanation of why a strategy was chosen
func (s *StrategyDeterminer) GetStrategyReason(flag domain.Flag) string {
	if !flag.Enabled {
		return "flag is disabled"
	}

	if len(flag.Segments) == 0 {
		return "no segments defined"
	}

	for _, segment := range flag.Segments {
		// Check rollout percentage
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			return "partial rollout requires consistent hashing"
		}

		// Check multiple distributions
		if len(segment.Distributions) > 1 {
			return "A/B testing with multiple variants requires consistent assignment"
		}

		// Check distribution percentage
		if len(segment.Distributions) == 1 && segment.Distributions[0].Percent < 100 {
			return "percentage-based distribution requires consistent hashing"
		}
	}

	return "100% deterministic based on constraints"
}

// AnalyzeFlag provides detailed analysis of flag evaluation characteristics
func (s *StrategyDeterminer) AnalyzeFlag(flag domain.Flag) FlagAnalysis {
	analysis := FlagAnalysis{
		FlagKey:  flag.Key,
		Enabled:  flag.Enabled,
		Strategy: s.Determine(flag),
		Reason:   s.GetStrategyReason(flag),
	}

	// Analyze segments
	for _, segment := range flag.Segments {
		segmentAnalysis := SegmentAnalysis{
			SegmentID:      segment.ID,
			RolloutPercent: segment.RolloutPercent,
			RequiresFlagr:  s.requiresFlagr(segment),
		}

		// Count constraints
		segmentAnalysis.ConstraintCount = len(segment.Constraints)

		// Analyze distributions
		segmentAnalysis.DistributionCount = len(segment.Distributions)
		if len(segment.Distributions) > 1 {
			segmentAnalysis.IsABTest = true
		}

		analysis.Segments = append(analysis.Segments, segmentAnalysis)
	}

	return analysis
}

// FlagAnalysis contains detailed analysis of a flag
type FlagAnalysis struct {
	FlagKey  string
	Enabled  bool
	Strategy domain.EvaluationStrategy
	Reason   string
	Segments []SegmentAnalysis
}

// SegmentAnalysis contains analysis of a single segment
type SegmentAnalysis struct {
	SegmentID         int64
	RolloutPercent    int
	ConstraintCount   int
	DistributionCount int
	IsABTest          bool
	RequiresFlagr     bool
}

// EstimatePerformance estimates the performance characteristics of evaluating this flag
func (s *StrategyDeterminer) EstimatePerformance(flag domain.Flag) PerformanceEstimate {
	estimate := PerformanceEstimate{
		Strategy: s.Determine(flag),
	}

	if estimate.Strategy == domain.StrategyLocal {
		// Local evaluation is very fast
		estimate.ExpectedLatency = "< 1ms"
		estimate.HTTPRequests = 0
		estimate.CacheHitRequired = true
	} else {
		// Remote evaluation requires HTTP call to Flagr
		estimate.ExpectedLatency = "50-200ms"
		estimate.HTTPRequests = 1
		estimate.CacheHitRequired = false
	}

	// Calculate complexity score
	estimate.ComplexityScore = s.calculateComplexity(flag)

	return estimate
}

// calculateComplexity calculates a complexity score for the flag
func (s *StrategyDeterminer) calculateComplexity(flag domain.Flag) int {
	score := 0

	// Base complexity
	score += len(flag.Segments) * 10

	for _, segment := range flag.Segments {
		// Constraints add complexity
		score += len(segment.Constraints) * 5

		// Distributions add complexity
		score += len(segment.Distributions) * 3

		// Partial rollout adds complexity
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			score += 20
		}
	}

	return score
}

// PerformanceEstimate contains performance estimates for flag evaluation
type PerformanceEstimate struct {
	Strategy         domain.EvaluationStrategy
	ExpectedLatency  string
	HTTPRequests     int
	CacheHitRequired bool
	ComplexityScore  int
}
