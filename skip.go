package conch

import (
	"context"
	"sync"
)

// Skip generates a stream by skipping the first count items from the
// input stream.
//
// Output stream is closed when the input stream is closed or context ctx is
// canceled.
func Skip[T any](
	ctx context.Context,
	inStream <-chan T,
	count int,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for i := 0; i < count; i++ {
			select {
			case <-ctx.Done():
				return
			case _, more := <-inStream:
				if !more {
					return
				}
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case e, more := <-inStream:
				if !more {
					return
				}
				select {
				case <-ctx.Done():
					return
				case outStream <- e:
				}
			}
		}
	}()

	return outStream
}

func SkipC[T any](
	count int,
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan T) {
		s := Skip(ctx, inStream, count)
		chain(ctx, wg, s)
	}
}
