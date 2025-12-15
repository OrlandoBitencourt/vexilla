package evaluator

import (
	"context"
	"fmt"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Evaluator defines the interface for flag evaluation
type Evaluator interface {
	// Evaluate evaluates a flag locally based on context
	Evaluate(ctx context.Context, flag domain.Flag, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error)

	// CanEvaluateLocally determines if a flag can be evaluated locally
	CanEvaluateLocally(flag domain.Flag) bool
}

// LocalEvaluator implements local flag evaluation using expression engine
type LocalEvaluator struct {
	programCache map[string]*vm.Program
}

// New creates a new local evaluator
func New() *LocalEvaluator {
	return &LocalEvaluator{
		programCache: make(map[string]*vm.Program),
	}
}

// Evaluate evaluates a flag locally
func (e *LocalEvaluator) Evaluate(ctx context.Context, flag domain.Flag, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	// Check if flag is enabled
	if !flag.Enabled {
		return &domain.EvaluationResult{
			FlagID:           flag.ID,
			FlagKey:          flag.Key,
			EvaluationReason: "flag disabled",
		}, nil
	}

	// No segments = return default (first variant)
	if len(flag.Segments) == 0 {
		return e.defaultResult(flag, "no segments"), nil
	}

	// Evaluate segments in rank order
	sortedSegments := flag.SortedSegments()

	for _, segment := range sortedSegments {
		// Check if segment matches
		matched, err := e.evaluateSegment(segment, evalCtx)
		if err != nil {
			return nil, domain.NewEvaluationError(flag.Key, "segment evaluation failed", err)
		}

		if !matched {
			continue
		}

		// Segment matched - return variant
		if len(segment.Distributions) == 0 {
			return e.defaultResult(flag, "no distributions in matched segment"), nil
		}

		// For local evaluation, we take the first distribution
		// (since we only handle 100% deterministic cases)
		dist := segment.Distributions[0]
		variant, found := flag.GetVariantByID(dist.VariantID)
		if !found {
			return nil, domain.NewEvaluationError(flag.Key, fmt.Sprintf("variant %d not found", dist.VariantID), nil)
		}

		return &domain.EvaluationResult{
			FlagID:            flag.ID,
			FlagKey:           flag.Key,
			SegmentID:         segment.ID,
			VariantID:         variant.ID,
			VariantKey:        variant.Key,
			VariantAttachment: variant.Attachment,
			EvaluationReason:  fmt.Sprintf("matched segment %d", segment.ID),
		}, nil
	}

	// No segments matched - return default
	return e.defaultResult(flag, "no segments matched"), nil
}

// evaluateSegment checks if a segment's constraints match the context
func (e *LocalEvaluator) evaluateSegment(segment domain.Segment, evalCtx domain.EvaluationContext) (bool, error) {
	// Empty constraints = always match
	if len(segment.Constraints) == 0 {
		return true, nil
	}

	// All constraints must match (AND logic)
	for _, constraint := range segment.Constraints {
		matched, err := e.evaluateConstraint(constraint, evalCtx)
		if err != nil {
			return false, err
		}

		if !matched {
			return false, nil
		}
	}

	return true, nil
}

// evaluateConstraint evaluates a single constraint
func (e *LocalEvaluator) evaluateConstraint(constraint domain.Constraint, evalCtx domain.EvaluationContext) (bool, error) {
	// Get property value from context
	propValue, exists := evalCtx.Context[constraint.Property]
	if !exists {
		return false, nil // Missing property = no match
	}

	// Evaluate based on operator
	switch constraint.Operator {
	case domain.OperatorEQ:
		return e.evaluateEquals(propValue, constraint.Value), nil

	case domain.OperatorNEQ:
		return !e.evaluateEquals(propValue, constraint.Value), nil

	case domain.OperatorIN:
		return e.evaluateIn(propValue, constraint.Value), nil

	case domain.OperatorNOTIN:
		return !e.evaluateIn(propValue, constraint.Value), nil

	case domain.OperatorMATCHES:
		return e.evaluateMatches(propValue, constraint.Value)

	case domain.OperatorLT:
		return e.evaluateCompare(propValue, constraint.Value, -1), nil

	case domain.OperatorLTE:
		cmp := e.evaluateCompare(propValue, constraint.Value, 0)
		return cmp, nil

	case domain.OperatorGT:
		return e.evaluateCompare(propValue, constraint.Value, 1), nil

	case domain.OperatorGTE:
		cmp := e.evaluateCompare(propValue, constraint.Value, 0)
		return cmp, nil

	default:
		return false, fmt.Errorf("unsupported operator: %s", constraint.Operator)
	}
}

// evaluateEquals checks equality
func (e *LocalEvaluator) evaluateEquals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// evaluateIn checks if value is in list
func (e *LocalEvaluator) evaluateIn(value interface{}, list interface{}) bool {
	listSlice, ok := list.([]interface{})
	if !ok {
		// Try to convert
		if strSlice, ok := list.([]string); ok {
			for _, item := range strSlice {
				if e.evaluateEquals(value, item) {
					return true
				}
			}
			return false
		}
		return false
	}

	for _, item := range listSlice {
		if e.evaluateEquals(value, item) {
			return true
		}
	}

	return false
}

// evaluateMatches checks regex match
func (e *LocalEvaluator) evaluateMatches(value interface{}, pattern interface{}) (bool, error) {
	// Build expression and evaluate
	exprStr := fmt.Sprintf(`value matches "%v"`, pattern)

	program, err := expr.Compile(exprStr, expr.Env(map[string]interface{}{"value": value}))
	if err != nil {
		return false, fmt.Errorf("failed to compile regex expression: %w", err)
	}

	result, err := expr.Run(program, map[string]interface{}{"value": value})
	if err != nil {
		return false, fmt.Errorf("failed to evaluate regex: %w", err)
	}

	matched, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("regex evaluation returned non-boolean: %T", result)
	}

	return matched, nil
}

// evaluateCompare performs numeric comparison
func (e *LocalEvaluator) evaluateCompare(a, b interface{}, expectedSign int) bool {
	// Convert to float64 for comparison
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	if expectedSign < 0 {
		return aFloat < bFloat
	} else if expectedSign > 0 {
		return aFloat > bFloat
	} else {
		return aFloat <= bFloat || aFloat >= bFloat
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

// defaultResult returns the default evaluation result
func (e *LocalEvaluator) defaultResult(flag domain.Flag, reason string) *domain.EvaluationResult {
	result := &domain.EvaluationResult{
		FlagID:           flag.ID,
		FlagKey:          flag.Key,
		EvaluationReason: reason,
	}

	// Try to get default value
	if len(flag.Segments) > 0 && len(flag.Segments[0].Distributions) > 0 {
		dist := flag.Segments[0].Distributions[0]
		if variant, found := flag.GetVariantByID(dist.VariantID); found {
			result.VariantID = variant.ID
			result.VariantKey = variant.Key
			result.VariantAttachment = variant.Attachment
		}
	} else if len(flag.Variants) > 0 {
		variant := flag.Variants[0]
		result.VariantID = variant.ID
		result.VariantKey = variant.Key
		result.VariantAttachment = variant.Attachment
	}

	return result
}

// CanEvaluateLocally determines if a flag can be safely evaluated locally
func (e *LocalEvaluator) CanEvaluateLocally(flag domain.Flag) bool {
	return flag.DetermineStrategy() == domain.StrategyLocal
}
