package evaluator

import (
	"context"
	"fmt"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
	"github.com/antonmedv/expr"
)

// Evaluator handles local flag evaluation using expressions
type Evaluator struct{}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// Evaluate evaluates a flag locally based on constraints
func (e *Evaluator) Evaluate(ctx context.Context, flag vexilla.Flag, evalCtx vexilla.EvaluationContext) (*vexilla.EvaluationResult, error) {
	// If flag is disabled, return default
	if !flag.Enabled {
		return &vexilla.EvaluationResult{
			FlagID:           flag.ID,
			FlagKey:          flag.Key,
			VariantKey:       "disabled",
			EvaluatedLocally: true,
		}, nil
	}

	// Build evaluation environment
	env := e.buildEnvironment(evalCtx)

	// Evaluate segments in rank order
	for _, segment := range flag.Segments {
		// Check if all constraints match
		if e.evaluateConstraints(segment.Constraints, env) {
			// Segment matched - return first distribution
			if len(segment.Distributions) > 0 {
				dist := segment.Distributions[0]
				return &vexilla.EvaluationResult{
					FlagID:            flag.ID,
					FlagKey:           flag.Key,
					SegmentID:         segment.ID,
					VariantID:         dist.VariantID,
					VariantKey:        dist.VariantKey,
					VariantAttachment: dist.VariantAttachment,
					EvaluatedLocally:  true,
				}, nil
			}
		}
	}

	// No segment matched - return nil/default
	return &vexilla.EvaluationResult{
		FlagID:           flag.ID,
		FlagKey:          flag.Key,
		VariantKey:       "default",
		EvaluatedLocally: true,
	}, nil
}

// buildEnvironment creates the evaluation environment from context
func (e *Evaluator) buildEnvironment(evalCtx vexilla.EvaluationContext) map[string]interface{} {
	env := make(map[string]interface{})

	// Add entity ID
	env["entityID"] = evalCtx.EntityID
	env["entityType"] = evalCtx.EntityType

	// Flatten context attributes
	for k, v := range evalCtx.Context {
		env[k] = v
	}

	return env
}

// evaluateConstraints checks if all constraints match
func (e *Evaluator) evaluateConstraints(constraints []vexilla.Constraint, env map[string]interface{}) bool {
	// No constraints = always match
	if len(constraints) == 0 {
		return true
	}

	// All constraints must match
	for _, constraint := range constraints {
		if !e.evaluateConstraint(constraint, env) {
			return false
		}
	}

	return true
}

// evaluateConstraint evaluates a single constraint
func (e *Evaluator) evaluateConstraint(constraint vexilla.Constraint, env map[string]interface{}) bool {
	value, exists := env[constraint.Property]
	if !exists {
		return false
	}

	switch constraint.Operator {
	case "EQ":
		return value == constraint.Value
	case "NEQ":
		return value != constraint.Value
	case "IN":
		if list, ok := constraint.Value.([]interface{}); ok {
			for _, item := range list {
				if value == item {
					return true
				}
			}
		}
		return false
	case "NOTIN":
		if list, ok := constraint.Value.([]interface{}); ok {
			for _, item := range list {
				if value == item {
					return false
				}
			}
			return true
		}
		return false
	case "MATCHES":
		// Use expr for regex matching
		if pattern, ok := constraint.Value.(string); ok {
			expression := fmt.Sprintf("matches(%q, %q)", value, pattern)
			program, err := expr.Compile(expression)
			if err != nil {
				return false
			}
			result, err := expr.Run(program, env)
			if err != nil {
				return false
			}
			if b, ok := result.(bool); ok {
				return b
			}
		}
		return false
	default:
		return false
	}
}
