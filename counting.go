package conch

import (
	"context"
	"sync"
)

func Counting[T any, R Integer](
	ctx context.Context,
	inStream <-chan T,
) <-chan R {
	outStream := make(chan R)

	go func() {
		var cnt R

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

type ChainFunc[T any] func(
	ctx context.Context,
	group *sync.WaitGroup,
	inStream <-chan T,
)

func CountingC[T any, R Integer](
	chain ChainFunc[R],
) ChainFunc[T] {
	return func(ctx context.Context, group *sync.WaitGroup, inStream <-chan T) {
		s := Counting[T, R](ctx, inStream)
		chain(ctx, group, s)
	}
}
