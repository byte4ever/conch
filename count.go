package conch

import (
	"context"
	"sync"
)

func Count[T any](
	ctx context.Context,
	inStream <-chan T,
) (count <-chan int) {
	var c int

	iCount := make(chan int)

	go func() {
		defer close(iCount)

		for range inStream {
			c++
		}

		select {
		case iCount <- c:
		case <-ctx.Done():
			return
		}

		iCount <- c
	}()

	return iCount
}

func CountC[T any](chain ChainFunc[int]) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Count(ctx, inStream))
	}
}
