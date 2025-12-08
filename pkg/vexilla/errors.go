package vexilla

import "fmt"

// ErrInvalidConfig represents a configuration validation error
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid config field '%s': %s", e.Field, e.Reason)
}

// ErrFlagNotFound indicates a flag was not found in cache
type ErrFlagNotFound struct {
	FlagKey string
}

func (e ErrFlagNotFound) Error() string {
	return fmt.Sprintf("flag not found: %s", e.FlagKey)
}

// ErrCircuitOpen indicates the circuit breaker is open
type ErrCircuitOpen struct {
	ConsecutiveFails int
}

func (e ErrCircuitOpen) Error() string {
	return fmt.Sprintf("circuit breaker is open after %d consecutive failures", e.ConsecutiveFails)
}

// ErrFlagrUnavailable indicates Flagr server is unavailable
type ErrFlagrUnavailable struct {
	Endpoint string
	Err      error
}

func (e ErrFlagrUnavailable) Error() string {
	return fmt.Sprintf("flagr unavailable at %s: %v", e.Endpoint, e.Err)
}
