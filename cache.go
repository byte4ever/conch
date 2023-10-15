package conch

import (
	"context"
	"sync"
)

type Cache[P, R any] interface {
	Get(ctx context.Context, key P) (R, bool)
	Store(ctx context.Context, key P, value R)
}

func CacheWriteInterceptor[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	wg.Add(1)

	go func() {
		defer wg.Done()
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

func CacheWriteInterceptors[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStreams ...<-chan Request[P, R],
) (outStreams []<-chan Request[P, R]) {
	outStreams = make([]<-chan Request[P, R], len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = CacheWriteInterceptor(
			ctx, wg, cache, inStream,
		)
	}

	return
}

func CacheWriteInterceptorC[P, R any](
	cache Cache[P, R],
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, CacheWriteInterceptor(ctx, wg, cache, inStream))
	}
}

func CacheWriteInterceptorsC[P, R any](
	cache Cache[P, R],
	chains ChainsFunc[Request[P, R]],
) ChainsFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan Request[P, R],
	) {
		chains(ctx, wg, CacheWriteInterceptors(ctx, wg, cache, inStreams...)...)
	}
}

func CacheReadInterceptor[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	wg.Add(1)

	go func() {
		defer wg.Done()
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

func CacheReadInterceptors[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStreams ...<-chan Request[P, R],
) (outStreams []<-chan Request[P, R]) {
	outStreams = make([]<-chan Request[P, R], len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = CacheReadInterceptor(ctx, wg, cache, inStream)
	}

	return
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
		chain(ctx, wg, CacheReadInterceptor(ctx, wg, cache, inStream))
	}
}

func CacheReadInterceptorsC[P, R any](
	cache Cache[P, R],
	chains ChainsFunc[Request[P, R]],
) ChainsFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan Request[P, R],
	) {
		chains(ctx, wg, CacheReadInterceptors(ctx, wg, cache, inStream...)...)
	}
}
