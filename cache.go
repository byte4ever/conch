package conch

import (
	"context"
	"sync"
)

func CacheWriteInterceptor[P Hashable, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			for r := range inStream {
				r.Chan <- ValErrorPair[R]{
					Err: ctx.Err(),
				}
			}
		}()
		defer close(outStream)

		valErrorChanPool := newValErrorChanPool[R](1000)

		for {
			select {
			case <-ctx.Done():
				return

			case req, more := <-inStream:
				if !more {
					return
				}

				cacheChan := valErrorChanPool.get()
				reqChan := req.Chan

				req.Chan = cacheChan

				select {
				case <-ctx.Done():
					reqChan <- ValErrorPair[R]{
						Err: ctx.Err(),
					}

				case outStream <- req:
					// intercept response channel and store to cache if no error
					// occurs
					wg.Add(1)

					go func() {
						defer wg.Done()
						defer valErrorChanPool.putBack(cacheChan)

						select {
						case <-ctx.Done():
							reqChan <- ValErrorPair[R]{
								Err: ctx.Err(),
							}

						case v, more := <-cacheChan:
							if !more {
								return
							}

							if v.Err == nil {
								cache.Store(ctx, req.P, v.V)
							}

							select {
							case <-ctx.Done():
								reqChan <- ValErrorPair[R]{
									Err: ctx.Err(),
								}
							case reqChan <- v:
							}
						}
					}()
				}
			}
		}
	}()

	return outStream
}

func CacheWriteInterceptors[P Hashable, R any](
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

func CacheWriteInterceptorC[P Hashable, R any](
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

func CacheWriteInterceptorsC[P Hashable, R any](
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

func CacheReadInterceptor[P Hashable, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	cache Cache[P, R],
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			for r := range inStream {
				r.Chan <- ValErrorPair[R]{
					Err: ctx.Err(),
				}
			}
		}()
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case req, more := <-inStream:
				if !more {
					return
				}

				if value, found := cache.Get(ctx, req.P); found {
					select {
					case <-ctx.Done():
						return
					case req.Chan <- ValErrorPair[R]{
						V: value,
					}:
					}
				} else {
					select {
					case <-ctx.Done():
						return

					case outStream <- req:
					}
				}
			}
		}
	}()

	return outStream
}

func CacheReadInterceptors[P Hashable, R any](
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

func CacheReadInterceptorC[P Hashable, R any](
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

func CacheReadInterceptorsC[P Hashable, R any](
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
