package conch

import (
	"context"
	"sync"
	"sync/atomic"
)

// Count counts how many elements are going through a single stream.
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

// CountsC counts how many elements are going through multiple streams.
func CountsC[T any](
	counter *atomic.Uint64,
	chain ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan T,
	) {
		for _, inStream := range inStreams {
			chain(ctx, wg, Count(ctx, counter, inStream))
		}
	}
}
