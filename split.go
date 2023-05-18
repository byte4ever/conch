package conch

import (
	"context"
	"sync"
)

func Split[T any](
	ctx context.Context,
	count int,
	inStream <-chan T,
) []<-chan T {
	streams := make([]chan T, count)
	for i := 0; i < count; i++ {
		streams[i] = make(chan T)
	}

	for i := 0; i < count; i++ {
		go func(inStream <-chan T, outStream chan T) {
			defer close(outStream)

			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-inStream:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					case outStream <- v:
					}
				}
			}
		}(inStream, streams[i])
	}

	outStreams := make([]<-chan T, count)
	for idx, s := range streams {
		outStreams[idx] = s
	}

	return outStreams
}

func SplitC[T any](out ...ChainFunc[T]) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		streams := Split(ctx, len(out), inStream)
		for idx, s := range streams {
			out[idx](ctx, wg, s)
		}
	}
}
