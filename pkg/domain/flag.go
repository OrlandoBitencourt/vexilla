package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// Flag represents a feature flag with its evaluation rules
type Flag struct {
	ID          int64
	Key         string
	Description string
	Enabled     bool
	Segments    []Segment
	Variants    []Variant
	UpdatedAt   time.Time

	// Metadata
	Tags               []Tag
	DataRecordsEnabled bool
}

// Tag represents a Flagr API tag
type Tag struct {
	Value string
}

// Segment represents a targeting segment
type Segment struct {
	ID             int64
	Rank           int
	Description    string
	RolloutPercent int // 0-100
	Constraints    []Constraint
	Distributions  []Distribution
}

// Constraint represents a targeting constraint
type Constraint struct {
	ID       int64
	Property string
	Operator Operator
	Value    interface{}
}

// Distribution represents variant distribution within a segment
type Distribution struct {
	ID        int64
	VariantID int64
	Percent   int // 0-100
}

// Variant represents a flag variant
type Variant struct {
	ID         int64
	Key        string
	Attachment map[string]json.RawMessage
}

// Operator represents constraint operators
type Operator string

const (
	OperatorEQ       Operator = "EQ"
	OperatorNEQ      Operator = "NEQ"
	OperatorLT       Operator = "LT"
	OperatorLTE      Operator = "LTE"
	OperatorGT       Operator = "GT"
	OperatorGTE      Operator = "GTE"
	OperatorIN       Operator = "IN"
	OperatorNOTIN    Operator = "NOTIN"
	OperatorMATCHES  Operator = "MATCHES"
	OperatorCONTAINS Operator = "CONTAINS"
)

// EvaluationStrategy determines how a flag should be evaluated
type EvaluationStrategy string

const (
	StrategyLocal  EvaluationStrategy = "local"  // Can be evaluated locally
	StrategyRemote EvaluationStrategy = "remote" // Requires Flagr for consistent bucketing
)

// Validate validates the flag configuration
func (f *Flag) Validate() error {
	if f.Key == "" {
		return NewValidationError("flag key cannot be empty")
	}

	// Flags sem segmentos são válidos
	if len(f.Segments) == 0 {
		return nil
	}

	for i, segment := range f.Segments {
		if err := segment.Validate(); err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
	}

	variantMap := make(map[int64]bool)
	for _, v := range f.Variants {
		variantMap[v.ID] = true
	}

	for i, seg := range f.Segments {
		for _, dist := range seg.Distributions {
			// Só validar VariantID se existirem variants definidos
			if len(f.Variants) > 0 {
				if !variantMap[dist.VariantID] {
					return NewValidationError(
						fmt.Sprintf("segment %d distribution references unknown variantID=%d", i, dist.VariantID),
					)
				}
			}
		}
	}

	return nil
}

// Validate validates the segment configuration
func (s *Segment) Validate() error {
	if s.RolloutPercent < 0 || s.RolloutPercent > 100 {
		return NewValidationError("segment rollout percent must be between 0 and 100")
	}

	if s.RolloutPercent > 0 && len(s.Distributions) == 0 {
		return NewValidationError("segment must have at least one distribution when rollout > 0")
	}

	total := 0
	for _, dist := range s.Distributions {
		if dist.Percent < 0 || dist.Percent > 100 {
			return NewValidationError("distribution percent must be between 0 and 100")
		}
		total += dist.Percent
	}

	if len(s.Distributions) > 0 && total != s.RolloutPercent {
		return NewValidationError(
			fmt.Sprintf("distribution percent sum %d must equal rolloutPercent %d", total, s.RolloutPercent),
		)
	}

	return nil
}

// DetermineStrategy determines if a flag can be evaluated locally
func (f *Flag) DetermineStrategy() EvaluationStrategy {
	if !f.Enabled {
		return StrategyLocal // Disabled flags are always local
	}

	if len(f.Segments) == 0 {
		return StrategyLocal // No segments = no evaluation needed
	}

	for _, segment := range f.Segments {
		// Partial rollout requires Flagr's consistent hashing
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			return StrategyRemote
		}

		// Multiple distributions (A/B test) requires Flagr
		if len(segment.Distributions) > 1 {
			return StrategyRemote
		}

		// Single distribution with < 100% requires Flagr
		if len(segment.Distributions) == 1 {
			dist := segment.Distributions[0]
			if dist.Percent < 100 {
				return StrategyRemote
			}
		}
	}

	// All segments are 100% deterministic
	return StrategyLocal
}

// GetVariantByID finds a variant by ID
func (f *Flag) GetVariantByID(id int64) (*Variant, bool) {
	for _, v := range f.Variants {
		if v.ID == id {
			return &v, true
		}
	}
	return nil, false
}

// GetDefaultValue returns the default value for a flag
// This is the attachment of the first variant in the first segment's first distribution
func (f *Flag) GetDefaultValue() interface{} {
	if !f.Enabled {
		return nil
	}

	if len(f.Segments) == 0 || len(f.Segments[0].Distributions) == 0 {
		return nil
	}

	firstDist := f.Segments[0].Distributions[0]
	if variant, found := f.GetVariantByID(firstDist.VariantID); found {
		return variant.Attachment
	}

	return nil
}

// SortedSegments returns segments sorted by rank
func (f *Flag) SortedSegments() []Segment {
	segments := make([]Segment, len(f.Segments))
	copy(segments, f.Segments)

	// Bubble sort by rank (small list, simple is fine)
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[i].Rank > segments[j].Rank {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}

	return segments
}
