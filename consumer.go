package conch

import (
	"context"
	"sync"
)

func Consumer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	f Doer[T],
	inStream <-chan T,
) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return

			case e, more := <-inStream:
				if !more {
					return
				}

				f(ctx, e)
			}
		}
	}()
}

func ConsumerPool[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	f Doer[T],
	inStream <-chan T,
) {
	for i := 0; i < count; i++ {
		Consumer(ctx, wg, f, inStream)
	}
}

func ConsumerC[T any](
	f Doer[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		Consumer(ctx, wg, f, inStream)
	}
}

func ConsumerPoolC[T any](
	count int,
	f Doer[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		ConsumerPool(ctx, wg, count, f, inStream)
	}
}
