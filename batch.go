package conch

import (
	"context"
	"sync"
)

// Batch groups items from the input channel into batches of the specified size.
// It runs in a goroutine until the input channel is closed or context is done.
// Batches are sent to the output channel, which is also returned by the
// function. The context ensures early termination when cancellation or timeout
// occurs. A sync.WaitGroup must be provided to manage goroutine execution.
func Batch[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	size int,
	inStream <-chan T,
) <-chan []T {
	outStream := make(chan []T)

	wg.Add(1)

	go func() {
		defer func() {
			close(outStream)
			wg.Done()
		}()

		curBatch := make([]T, 0, size)

		for {
			select {
			case <-ctx.Done():
				return

			case elem, more := <-inStream:
				if !more {
					if len(curBatch) > 0 {
						select {
						case <-ctx.Done():
							return
						case outStream <- curBatch:
						}
					}

					return
				}

				if len(curBatch) == size {
					select {
					case <-ctx.Done():
						return
					case outStream <- curBatch:
						curBatch = make([]T, 0, size)
					}
				} else {
					curBatch = append(curBatch, elem)
				}
			}
		}
	}()

	return outStream
}

// BatchC creates a chainable function for batch processing of elements.
// It batches the input stream into slices of a specified size and chains
// the batched output to the next processing step.
func BatchC[T any](
	size int,
	chain ChainFunc[[]T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(
			ctx,
			wg,
			Batch(
				ctx,
				wg,
				size,
				inStream,
			),
		)
	}
}

// Batchs concurrently batches multiple input streams into slices of a given
// size. The function returns a slice of output channels, each containing the
// batched results. The batching process respects the given context and
// synchronizes using WaitGroup.
func Batchs[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	size int,
	inStreams ...<-chan T,
) (outStreams []<-chan []T) {
	outStreams = make(
		[]<-chan []T,
		0,
		len(inStreams),
	)

	for _, inStream := range inStreams {
		outStreams = append(
			outStreams,
			Batch(
				ctx,
				wg,
				size,
				inStream,
			),
		)
	}

	return
}

// BatchesC transforms a ChainsFunc to batch elements from input channels. It
// outputs batches of size `size` and processes them using provided ChainsFunc.
func BatchesC[T any](
	size int,
	chains ChainsFunc[[]T],
) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan T,
	) {
		chains(ctx, wg, Batchs(
			ctx,
			wg,
			size,
			inStream...,
		)...)
	}
}
