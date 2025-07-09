package conch

import (
	"context"
	"sync"
)

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
