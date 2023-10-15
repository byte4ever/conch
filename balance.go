package conch

import (
	"context"
	"sync"
)

func Balance[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	inStream <-chan T,
) []<-chan T {
	outStream := make([]chan T, count)
	for i := 0; i < count; i++ {
		outStream[i] = make(chan T)
	}

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(s chan T) {
			defer wg.Done()
			defer close(s)

			for t := range inStream {
				select {
				case <-ctx.Done():
					return
				case s <- t:
				}
			}
		}(outStream[i])
	}

	iOutStream := make([]<-chan T, count)
	for i, cc := range outStream {
		iOutStream[i] = cc
	}

	return iOutStream
}

func BalanceC[T any](
	count int,
	chains ChainsFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream <-chan T,
	) {
		chains(ctx, wg, Balance[T](ctx, wg, count, inStream)...)
	}
}
