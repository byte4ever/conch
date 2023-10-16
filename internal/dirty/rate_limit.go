package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
	"github.com/byte4ever/conch/internal/ratelimit"
)

func RateLimit[T any](
	ctx context.Context,
	inStream <-chan T,
	ratePerSecond int,
	option ...ratelimit.Option,
) <-chan T {
	outStream := make(chan T)

	rl := ratelimit.New(ratePerSecond, option...)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case <-rl.Take(ctx):
				select {
				case val, more := <-inStream:
					if !more {
						return
					}

					select {
					case outStream <- val:
					case <-ctx.Done():
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func RateLimitC[T any](
	ratePerSecond int,
	chain dirty.ChainFunc[T],
	option ...ratelimit.Option,
) dirty.ChainFunc[T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan T) {
		chain(ctx, wg, RateLimit(ctx, inStream, ratePerSecond, option...))
	}
}
