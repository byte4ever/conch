package conch

import (
	"context"
	"sync"
)

// Buffer chains a buffered channel to inStream.
// This may help inStreams writer to be less frequently in a blocking state.
//
// Keep in mind that as soon as buffer is full it behave exactly as if it
// wasn't buffered at all.
func Buffer[T any](
	ctx context.Context,
	inStream <-chan T,
	size int,
) <-chan T {
	outStream := make(chan T, size)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case item, more := <-inStream:
				if !more {
					return
				}

				select {
				case <-ctx.Done():
					return
				case outStream <- item:
				}
			}
		}
	}()

	return outStream
}

func BufferC[T any](size int, chain ChainFunc[T]) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Buffer(ctx, inStream, size))
	}
}
