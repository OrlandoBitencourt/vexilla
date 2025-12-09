package vexilla

import (
	"context"

	"github.com/antonmedv/expr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Evaluator handles flag rule evaluation using expr
type Evaluator struct {
	tracer trace.Tracer
}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{
		tracer: otel.Tracer(tracerName),
	}
}

// Evaluate evaluates flag rules against the context
func (e *Evaluator) Evaluate(ctx context.Context, flag Flag, evalCtx EvaluationContext) interface{} {
	_, span := e.tracer.Start(ctx, "evaluator.evaluate")
	defer span.End()

	env := map[string]interface{}{
		"user_id": evalCtx.UserID,
		"user":    evalCtx.Attributes,
	}

	for k, v := range evalCtx.Attributes {
		env[k] = v
	}

	for i, rule := range flag.Rules {
		if rule.Condition == "" {
			continue
		}

		program, err := expr.Compile(rule.Condition, expr.Env(env))
		if err != nil {
			span.AddEvent("rule_compilation_failed", trace.WithAttributes(
				attribute.Int("rule.index", i),
				attribute.String("error", err.Error()),
			))
			continue
		}

		output, err := expr.Run(program, env)
		if err != nil {
			span.AddEvent("rule_evaluation_failed", trace.WithAttributes(
				attribute.Int("rule.index", i),
				attribute.String("error", err.Error()),
			))
			continue
		}

		if result, ok := output.(bool); ok && result {
			span.SetAttributes(
				attribute.Int("matched_rule.index", i),
				attribute.String("matched_rule.condition", rule.Condition),
			)
			return rule.Value
		}
	}

	return flag.Default
}
