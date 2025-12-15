package vexilla

import (
	"fmt"
)

// Error types that may be returned by Vexilla operations.

// EvaluationError represents an error during flag evaluation.
type EvaluationError struct {
	FlagKey string
	Reason  string
	Err     error
}

func (e *EvaluationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("evaluation error for flag %s: %s: %v", e.FlagKey, e.Reason, e.Err)
	}
	return fmt.Sprintf("evaluation error for flag %s: %s", e.FlagKey, e.Reason)
}

func (e *EvaluationError) Unwrap() error {
	return e.Err
}

// NotFoundError indicates a flag was not found.
type NotFoundError struct {
	FlagKey string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("flag not found: %s", e.FlagKey)
}

// CircuitOpenError indicates the circuit breaker is open.
type CircuitOpenError struct {
	Message string
}

func (e *CircuitOpenError) Error() string {
	return fmt.Sprintf("circuit breaker is open: %s", e.Message)
}

// ConfigError indicates invalid configuration.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error [%s]: %s", e.Field, e.Message)
}
