// Package cbreaker provides circuit breaker functionality for stream processing.
// This file defines configuration types and options for circuit breaker behavior.
package cbreaker

import (
	"time"
)

type (
	// config holds the internal configuration parameters for circuit breaker behavior.
	// It defines thresholds and timeouts that control when the circuit breaker opens,
	// closes, or enters half-open state.
	config struct {
		nbFailureToOpen  *int           // Number of consecutive failures before opening the circuit
		nbSuccessToClose *int           // Number of consecutive successes needed to close the circuit
		halfOpenTimeout  *time.Duration // Duration to wait before allowing test requests in half-open state
	}

	// Option defines the interface for configuring circuit breaker options.
	// Implementations apply specific configuration settings to the circuit breaker.
	Option interface {
		apply(conf *config) error
	}
)

// ref is a utility function that returns a pointer to any value.
// It simplifies creating pointers to literals for configuration fields.
func ref[T any](v T) *T {
	return &v
}

var (
	// defaultConfig provides the default configuration values for circuit breaker.
	// These values represent reasonable defaults for most use cases:
	// - 4 failures trigger circuit opening
	// - 2 successes trigger circuit closing
	// - 5 second timeout for half-open state
	defaultConfig = config{ //nolint:gochecknoglobals // thats default
		nbFailureToOpen:  ref(4),
		nbSuccessToClose: ref(2),
		halfOpenTimeout:  ref(5 * time.Second),
	}
)
