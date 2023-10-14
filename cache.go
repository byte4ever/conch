package conch

import (
	"context"
	"sync"
)

type Cache[P, R any] interface {
	Get(ctx context.Context, key P) (R, bool)
	Store(ctx context.Context, key P, value R)
}

func CacheInterceptor[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	go func() {
		defer close(outStream)

		valErrorChanPool := newValErrorChanPool[R](maxCapacity)

	again:
		select {
		case <-ctx.Done():
			return
		case req, more := <-inStream:
			if !more {
				return
			}

			value, found := cache.Get(ctx, req.P)
			if found {
				select {
				case <-ctx.Done():
					return
				case req.Chan <- ValErrorPair[R]{
					V: value,
				}:
					goto again
				}
			}

			cacheChan := valErrorChanPool.get()

			select {
			case <-ctx.Done():
				return

			case outStream <- Request[P, R]{
				P:    req.P,
				Ctx:  ctx,
				Chan: cacheChan,
			}:
				// intercept response channel and store to cache if no error occurs
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer valErrorChanPool.putBack(cacheChan)
					select {
					case v, more := <-cacheChan:
						if !more {
							return
						}

						if v.Err == nil {
							cache.Store(ctx, req.P, v.V)
						}

						select {
						case <-req.Ctx.Done():
							return
						case req.Chan <- v:
						}
					}
				}()

				goto again
			}
		}
	}()

	return outStream
}

func CacheInterceptorC[P, R any](
	cache Cache[P, R],
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, CacheInterceptor(ctx, wg, cache, inStream))
	}
}

func CacheReadInterceptor[P, R any](
	ctx context.Context,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	go func() {
		defer close(outStream)

	again:
		select {
		case <-ctx.Done():
			return
		case req, more := <-inStream:
			if !more {
				return
			}

			value, found := cache.Get(ctx, req.P)
			if found {
				select {
				case <-ctx.Done():
					return
				case req.Chan <- ValErrorPair[R]{
					V:   value,
					Err: nil,
				}:
				}

				goto again
			}

			select {
			case <-ctx.Done():
				return

			case outStream <- req:
				goto again
			}
		}
	}()

	return outStream
}

func CacheReadInterceptorC[P, R any](
	cache Cache[P, R],
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, CacheReadInterceptor(ctx, cache, inStream))
	}
}

func CacheReadInterceptorPool[P, R any](
	ctx context.Context,
	count int,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	streams := make([]<-chan Request[P, R], count)
	for i := 0; i < count; i++ {
		streams[i] = CacheReadInterceptor(ctx, cache, inStream)
	}

	return FanIn(ctx, streams...)
}

func CacheReadInterceptorPoolC[P, R any](
	count int,
	cache Cache[P, R],
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(
			ctx,
			wg,
			CacheReadInterceptorPool(
				ctx,
				count,
				cache,
				inStream,
			),
		)
	}
}
