// Package conch provides stream processing utilities for Go applications.
// This file implements counting utilities for tracking stream elements.
package conch

import (
	"context"
	"sync"
	"sync/atomic"
)

// Count starts a goroutine that counts elements flowing through a stream.
// It reads from inStream, counts each element, and forwards them to the output channel.
// The count is stored in the provided atomic.Uint64 counter for thread-safe access.
func Count[T any](
	ctx context.Context,
	counter *atomic.Uint64,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case in, more := <-inStream:
				if !more {
					return
				}

				select {
				case outStream <- in:
					counter.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

// CountsC creates a chainable function for counting elements across multiple streams.
// It wraps the Count function to work with the ChainsFunc interface, allowing it to be
// composed with other stream operations in a processing pipeline.
func CountsC[T any](
	counter *atomic.Uint64,
	chain ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan T,
	) {
		// Apply counting to each input stream individually
		for _, inStream := range inStreams {
			chain(ctx, wg, Count(ctx, counter, inStream))
		}
	}
}
