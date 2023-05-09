package conch

import (
	"context"
	"sync"
)

// Keep generates a stream by keeping the first count items from the
// input stream.
//
// Output stream is closed when the input stream is closed or context ctx is
// canceled.
func Keep[T any](
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

		for {
			select {
			case <-ctx.Done():
				return
			case _, more := <-inStream:
				if !more {
					return
				}
			}
		}
	}()

	return outStream
}

func KeepC[T any](
	count int,
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan T) {
		s := Keep(ctx, inStream, count)
		chain(ctx, wg, s)
	}
}
