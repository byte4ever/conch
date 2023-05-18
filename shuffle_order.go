package conch

import (
	"context"
	"math/rand"
	"sync"
)

func ShuffleOrder[T any](
	ctx context.Context,
	maxBufferSize int,
	rnd *rand.Rand,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		buf := make([]T, 0, maxBufferSize)

		defer func() {
			// try to flush buffer
			if len(buf) > 0 {
				for _, i := range buf {
					select {
					case <-ctx.Done():
						return
					case outStream <- i:
					}
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case item, more := <-inStream:
				if !more {
					return
				}

				switch len(buf) {
				case maxBufferSize:
					// reach buffer saturation item MUST go out
					buf[maxBufferSize-1], item = item, buf[maxBufferSize-1]
					rnd.Shuffle(
						len(buf), func(i, j int) {
							buf[i], buf[j] = buf[j], buf[i]
						},
					)
					select {
					case <-ctx.Done():
						return
					case outStream <- item:
					}

				default:
					buf = append(buf, item)
					rnd.Shuffle(
						len(buf), func(i, j int) {
							buf[i], buf[j] = buf[j], buf[i]
						},
					)

					item = buf[len(buf)-1]

					select {
					case <-ctx.Done():
						return
					case outStream <- item:
						buf = buf[:len(buf)-1]
					default:
					}
				}
			}
		}
	}()

	return outStream
}

func ShuffleOrderC[T any](
	maxBufferSize int,
	rnd *rand.Rand,
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, ShuffleOrder(ctx, maxBufferSize, rnd, inStream))
	}
}
