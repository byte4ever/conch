package conch

import (
	"context"
	"sync"
)

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
