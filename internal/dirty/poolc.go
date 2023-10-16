package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

func PoolC[T any](
	count int,
	chain dirty.ChainFunc[T],
) dirty.ChainFunc[T] {
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
