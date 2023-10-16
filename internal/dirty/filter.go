package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

func Filter[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	filter dirty.FilterFunc[T],
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case in, more := <-inStream:
				if !more {
					return
				}

				if filter(ctx, in) && ctx.Err() == nil {
					select {
					case outStream <- in:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return outStream
}

func Filters[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	filter dirty.FilterFunc[T],
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	outStreams = make([]<-chan T, len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = Filter(ctx, wg, filter, inStream)
	}

	return
}

func FilterC[T any](
	filter dirty.FilterFunc[T],
	chain dirty.ChainFunc[T],
) dirty.ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Filter(ctx, wg, filter, inStream))
	}
}

func FiltersC[T any](
	filter dirty.FilterFunc[T],
	chains dirty.ChainsFunc[T],
) dirty.ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan T,
	) {
		chains(ctx, wg, Filters(ctx, wg, filter, inStreams...)...)
	}
}
