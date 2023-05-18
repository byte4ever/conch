package conch

import (
	"context"
	"sync"
)

func Counting[T any](
	ctx context.Context,
	inStream <-chan T,
) <-chan int {
	outStream := make(chan int)

	go func() {
		var cnt int

		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case _, more := <-inStream:
				if !more {
					return
				}

				select {
				case outStream <- cnt:
					cnt++
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func CountingC[T any](
	chain ChainFunc[int],
) ChainFunc[T] {
	return func(
		ctx context.Context, group *sync.WaitGroup, inStream <-chan T,
	) {
		s := Counting(ctx, inStream)
		chain(ctx, group, s)
	}
}
