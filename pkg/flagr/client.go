package flagr

import (
	"context"

	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
)

// Client defines the interface for Flagr communication (HTTP, mock, etc.)
type Client interface {
	// GetAllFlags fetches all flags with full details
	GetAllFlags(ctx context.Context) ([]domain.Flag, error)

	// GetFlag fetches a single flag by ID with full details
	GetFlag(ctx context.Context, flagID int64) (*domain.Flag, error)

	// EvaluateFlag remotely evaluates a flag using Flagr
	EvaluateFlag(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error)

	// HealthCheck checks if Flagr is reachable
	HealthCheck(ctx context.Context) error
}
