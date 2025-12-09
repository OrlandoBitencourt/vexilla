package types

import "context"

// FlagEvaluator defines the interface for flag evaluation
type FlagEvaluator interface {
	CanEvaluateLocally(flag *Flag) bool
	Evaluate(flag *Flag, ctx EvaluationContext) (bool, error)
}

// FlagrClient defines the interface for Flagr communication
type FlagrClient interface {
	GetFlag(ctx context.Context, flagKey string) (*Flag, error)
	EvaluateFlag(ctx context.Context, flagKey string, evalCtx EvaluationContext) (bool, error)
	ListFlags(ctx context.Context) ([]*Flag, error)
}

// FlagCache defines the interface for caching
type FlagCache interface {
	GetFlag(flagKey string) (*Flag, bool)
	SetFlag(flag *Flag) error
	SetFlags(flags []*Flag) error
	InvalidateFlag(flagKey string)
	Clear()
	GetStats() Stats
	HitRate() float64
	ListKeys() []string
	CanEvaluateLocally(flagKey string) bool
	EvaluateLocal(flagKey string, ctx EvaluationContext) (bool, error)
	Close() error
	WarmUp(ctx context.Context, flags []*Flag) error
}
