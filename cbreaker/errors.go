// Package cbreaker provides circuit breaker functionality for stream processing.
// This file defines error variables for circuit breaker configuration validation.
package cbreaker

import (
	"errors"
)

var (
	// ErrInvalidFailureToOpen is returned when the failure threshold is invalid.
	// This error occurs when attempting to set a zero or negative value for the
	// number of failures required to open the circuit breaker.
	ErrInvalidFailureToOpen = errors.New("must be non zero positive integer")

	// ErrInvalidSuccessToClose is returned when the success threshold is invalid.
	// This error occurs when attempting to set a zero or negative value for the
	// number of successes required to close the circuit breaker.
	ErrInvalidSuccessToClose = errors.New("must be non zero positive integer")

	// ErrInvalidOpenTimeout is returned when the timeout duration is invalid.
	// This error occurs when attempting to set a zero or negative duration for
	// the half-open timeout period.
	ErrInvalidOpenTimeout = errors.New("must be non zero positive duration")
)
