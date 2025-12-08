package evaluator

import "github.com/OrlandoBitencourt/vexilla/pkg/vexilla"

// Strategy determines if a flag can be evaluated locally
type Strategy string

const (
	StrategyLocal  Strategy = "local"  // Can evaluate locally
	StrategyRemote Strategy = "remote" // Requires Flagr
)

// Determiner determines evaluation strategy for flags
type Determiner struct{}

// NewDeterminer creates a new strategy determiner
func NewDeterminer() *Determiner {
	return &Determiner{}
}

// Determine determines if a flag can be evaluated locally
func (d *Determiner) Determine(flag vexilla.Flag) Strategy {
	// Disabled flags always return default (local)
	if !flag.Enabled {
		return StrategyLocal
	}

	// No segments = always default (local)
	if len(flag.Segments) == 0 {
		return StrategyLocal
	}

	// Check each segment
	for _, segment := range flag.Segments {
		// Partial rollout requires Flagr for consistent bucketing
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			return StrategyRemote
		}

		// Multiple distributions = A/B test = needs Flagr
		if len(segment.Distributions) > 1 {
			return StrategyRemote
		}

		// Single distribution with < 100% needs Flagr
		if len(segment.Distributions) == 1 {
			if segment.Distributions[0].Percent < 100 {
				return StrategyRemote
			}
		}
	}

	// All segments are 100% rollout = deterministic = local
	return StrategyLocal
}

// CanEvaluateLocally checks if a flag can be safely evaluated locally
func (d *Determiner) CanEvaluateLocally(flag vexilla.Flag) bool {
	return d.Determine(flag) == StrategyLocal
}
