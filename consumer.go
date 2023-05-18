package conch

import (
	"context"
	"runtime/debug"
	"sync"
)

func intercept[T any](
	logger Logger,
	f Doer[T],
) Doer[T] {
	return func(ctx context.Context, t T) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(
					"intercept panic",
					map[string]any{
						"error": r,
						"stack": string(debug.Stack()),
					},
				)
			}
		}()

		f(ctx, t)
	}
}

func Consumer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	f Doer[T],
	inStream <-chan T,
) {
	consumer(ctx, wg, logger, f, inStream)
}

func consumer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	l Logger,
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

				intercept(l, f)(ctx, e)
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
