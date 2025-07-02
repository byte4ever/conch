package conch

import (
	"context"
	"sync"
)

func Sink[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
	inStream <-chan *T,
) {
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-inStream:
				if !more {
					return
				}
				f(ctx, v)
				poolReceiver.PutBack(v)
			}
		}
	}()
}

func SinkC[T any](
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
) ChainFunc[*T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan *T) {
		Sink(
			ctx,
			wg,
			poolReceiver,
			f,
			inStream,
		)
	}
}

func Sinks[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
	inStreams ...<-chan *T,
) {
	for _, inStream := range inStreams {
		Sink(ctx, wg, poolReceiver, f, inStream)
	}
}

func SinksC[T any](
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
) ChainsFunc[*T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan *T,
	) {
		Sinks(
			ctx,
			wg,
			poolReceiver,
			f,
			inStream...,
		)
	}
}
