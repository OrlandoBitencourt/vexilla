package client

import (
	"context"

	"github.com/OrlandoBitencourt/vexilla/pkg/vexilla"
)

// MockClient is a mock Flagr client for testing
type MockClient struct {
	GetFlagsFunc       func(ctx context.Context) ([]vexilla.Flag, error)
	PostEvaluationFunc func(ctx context.Context, flagKey string, evalCtx vexilla.EvaluationContext) (*vexilla.EvaluationResult, error)
	HealthFunc         func(ctx context.Context) error
}

func (m *MockClient) GetFlags(ctx context.Context) ([]vexilla.Flag, error) {
	if m.GetFlagsFunc != nil {
		return m.GetFlagsFunc(ctx)
	}
	return []vexilla.Flag{}, nil
}

func (m *MockClient) PostEvaluation(ctx context.Context, flagKey string, evalCtx vexilla.EvaluationContext) (*vexilla.EvaluationResult, error) {
	if m.PostEvaluationFunc != nil {
		return m.PostEvaluationFunc(ctx, flagKey, evalCtx)
	}
	return &vexilla.EvaluationResult{FlagKey: flagKey, VariantKey: "mock"}, nil
}

func (m *MockClient) Health(ctx context.Context) error {
	if m.HealthFunc != nil {
		return m.HealthFunc(ctx)
	}
	return nil
}
