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
	wg *sync.WaitGroup,
	size int,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T, size)

	wg.Add(1)

	go func() {
		defer wg.Done()
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

func Buffers[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	size int,
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	outStreams = make([]<-chan T, len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = Buffer(ctx, wg, size, inStream)
	}

	return
}

func BufferC[T any](size int, chain ChainFunc[T]) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Buffer(ctx, wg, size, inStream))
	}
}

func BuffersC[T any](size int, chains ChainsFunc[T]) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan T,
	) {
		chains(ctx, wg, Buffers(ctx, wg, size, inStreams...)...)
	}
}
