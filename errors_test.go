package vexilla

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEvaluationError_Error tests EvaluationError formatting
func TestEvaluationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *EvaluationError
		wantText string
	}{
		{
			name: "with wrapped error",
			err: &EvaluationError{
				FlagKey: "test-flag",
				Reason:  "constraint mismatch",
				Err:     errors.New("invalid constraint"),
			},
			wantText: "evaluation error for flag test-flag: constraint mismatch: invalid constraint",
		},
		{
			name: "without wrapped error",
			err: &EvaluationError{
				FlagKey: "test-flag",
				Reason:  "not found",
				Err:     nil,
			},
			wantText: "evaluation error for flag test-flag: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.wantText, got)
		})
	}
}

// TestEvaluationError_Unwrap tests error unwrapping
func TestEvaluationError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	evalErr := &EvaluationError{
		FlagKey: "test-flag",
		Reason:  "failed",
		Err:     innerErr,
	}

	unwrapped := evalErr.Unwrap()
	assert.Equal(t, innerErr, unwrapped)

	// Test nil case
	nilErr := &EvaluationError{
		FlagKey: "test-flag",
		Reason:  "failed",
		Err:     nil,
	}
	assert.Nil(t, nilErr.Unwrap())
}

// TestNotFoundError_Error tests NotFoundError formatting
func TestNotFoundError_Error(t *testing.T) {
	err := &NotFoundError{
		FlagKey: "missing-flag",
	}

	got := err.Error()
	assert.Equal(t, "flag not found: missing-flag", got)
}

// TestCircuitOpenError_Error tests CircuitOpenError formatting
func TestCircuitOpenError_Error(t *testing.T) {
	err := &CircuitOpenError{
		Message: "too many failures",
	}

	got := err.Error()
	assert.Equal(t, "circuit breaker is open: too many failures", got)
}

// TestConfigError_Error tests ConfigError formatting
func TestConfigError_Error(t *testing.T) {
	err := &ConfigError{
		Field:   "endpoint",
		Message: "cannot be empty",
	}

	got := err.Error()
	assert.Equal(t, "configuration error [endpoint]: cannot be empty", got)
}
