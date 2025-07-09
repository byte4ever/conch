package conch

import (
	"context"
	"sync"
)

// Unbatch converts a stream of slices into a flat stream of individual
// elements. Processes slices received via the input channel and emits elements
// sequentially. Supports context cancellation and waits for completion via
// WaitGroup. Closes the output channel when all elements are processed or
// context is done.
func Unbatch[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan []T,
) <-chan T {
	outStream := make(chan T)

	wg.Add(1)

	go func() {
		defer func() {
			close(outStream)
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case slice, more := <-inStream:
				if !more {
					return
				}

				for _, t := range slice {
					select {
					case <-ctx.Done():
						return
					case outStream <- t:
					}
				}
			}
		}
	}()

	return outStream
}

// UnbatchC combines Unbatch with a ChainFunc for processing slice streams. It
// serializes elements from slices and passes results to the next ChainFunc.
func UnbatchC[T any](
	chain ChainFunc[T],
) ChainFunc[[]T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan []T,
	) {
		chain(
			ctx,
			wg,
			Unbatch(
				ctx,
				wg,
				inStream,
			),
		)
	}
}

// Unbatches transforms multiple channels of slices into channels of individual
// elements. It utilizes the Unbatch function for each input channel and returns
// a slice of output channels. Supports context cancellation and ensures proper
// synchronization via WaitGroup.
func Unbatches[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStreams ...<-chan []T,
) (outStreams []<-chan T) {
	outStreams = make(
		[]<-chan T,
		0,
		len(inStreams),
	)

	for _, inStream := range inStreams {
		outStreams = append(
			outStreams,
			Unbatch(
				ctx,
				wg,
				inStream,
			),
		)
	}

	return
}

// UnbatchesC transforms a ChainsFunc of slices into a ChainsFunc of elements.
// It uses the Unbatches function to flatten incoming streams of slices.
// Supports context handling and WaitGroup synchronization.
func UnbatchesC[T any](
	chains ChainsFunc[T],
) ChainsFunc[[]T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan []T,
	) {
		chains(
			ctx,
			wg,
			Unbatches(
				ctx,
				wg,
				inStream...,
			)...,
		)
	}
}
