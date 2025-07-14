// Package conch provides stream processing utilities for Go applications.
// This file implements caching functionality for stream processing elements.
package conch

import (
	"context"

	"github.com/byte4ever/conch/domain"
)

// Cachable represents an element that can be cached and processed.
// It combines the ability to be processed with parameter/result types
// and to retrieve the cached value. This interface is used for elements
// that need to maintain their computed results for reuse.
type Cachable[Param domain.Hashable, Result any] interface {
	domain.Processable[Param, Result]
	GetValue() Result
}

// ProcessorCacheFunc defines a function that processes cachable elements.
// It takes a context for cancellation and a cachable element to process.
// The function is responsible for updating the element's cached value
// or setting an error if processing fails.
type ProcessorCacheFunc[T Cachable[Param, Result], Param domain.Hashable, Result any] func(
	ctx context.Context,
	elem T,
)
