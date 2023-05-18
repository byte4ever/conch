package conch

import (
	"context"
	"sync"
)

func Filter[T any](
	ctx context.Context,
	inStream <-chan T,
	filter FilterFunc[T],
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

func FilterC[T any](
	filter FilterFunc[T],
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Filter(ctx, inStream, filter))
	}
}
