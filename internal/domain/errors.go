package domain

import (
	"fmt"
)

// -----------------------------
// CircuitOpenError
// -----------------------------

type CircuitOpenError struct {
	Message string
}

func NewCircuitOpenError(message string) *CircuitOpenError {
	return &CircuitOpenError{Message: message}
}

func (e *CircuitOpenError) Error() string {
	return fmt.Sprintf("circuit open: %s", e.Message)
}

func IsCircuitOpen(err error) bool {
	_, ok := err.(*CircuitOpenError)
	return ok
}

// -----------------------------
// NotFoundError
// -----------------------------

type NotFoundError struct {
	Resource string
	Key      string
}

func NewNotFoundError(resource, key string) *NotFoundError {
	return &NotFoundError{Resource: resource, Key: key}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Key)
}

func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// -----------------------------
// EvaluationError
// -----------------------------

type EvaluationError struct {
	FlagKey string
	Reason  string
	Err     error
}

func NewEvaluationError(flagKey, reason string, err error) *EvaluationError {
	return &EvaluationError{
		FlagKey: flagKey,
		Reason:  reason,
		Err:     err,
	}
}

func (e *EvaluationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("evaluation error on flag %s: %s: %v", e.FlagKey, e.Reason, e.Err)
	}
	return fmt.Sprintf("evaluation error on flag %s: %s", e.FlagKey, e.Reason)
}

func (e *EvaluationError) Unwrap() error {
	return e.Err
}

func IsEvaluationError(err error) bool {
	_, ok := err.(*EvaluationError)
	return ok
}

// -----------------------------
// ValidationError
// -----------------------------

type ValidationError struct {
	Message string
	Cause   error
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}

func NewValidationErrorWithCause(message string, cause error) *ValidationError {
	return &ValidationError{
		Message: message,
		Cause:   cause,
	}
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("validation error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
