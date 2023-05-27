package conch

import (
	"context"
	"sync"
)

func PoolC[T any](
	count int,
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		for i := 0; i < count; i++ {
			chain(ctx, wg, inStream)
		}
	}
}
