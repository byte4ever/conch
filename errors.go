// Package conch provides stream processing utilities for Go applications.
// This file contains custom error types for more informative error handling.
package conch

import (
	"fmt"
)

// PanicError represents an error that includes stack trace information.
// It can be used to provide more context when recovering from panics.
type PanicError struct {
	err   error  // The original error.
	stack string // The stack trace captured at the time of the panic.
}

// Error implements the error interface for PanicError, returning a formatted error string.
func (e *PanicError) Error() string {
	return fmt.Sprintf("error:%v\nstack:\n%s", e.err, e.stack)
}
