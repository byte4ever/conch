package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

type BreakerEngine interface {
	IsOpen() bool
	ReportFailure()
	ReportSuccess()
}

func Breaker[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	engine BreakerEngine,
	breakerError error,
	inStream <-chan dirty.Request[P, R],
) <-chan dirty.Request[P, R] {
	outStream := make(chan dirty.Request[P, R])

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(outStream)

		chanPool := newValErrorChanPool[R](maxCapacity)

		for {
			select {
			case <-ctx.Done():
				return
			case req, more := <-inStream:
				if !more {
					return
				}

				if engine.IsOpen() {
					select {
					case <-ctx.Done():
						return
					case req.Chan <- dirty.ValErrorPair[R]{Err: breakerError}:
					}

					continue
				}

				trackedChan := chanPool.get()

				wg.Add(1)

				go func() {
					defer wg.Done()
					defer chanPool.putBack(trackedChan)
					select {
					case v := <-trackedChan:
						if v.Err != nil {
							engine.ReportFailure()
						} else {
							engine.ReportSuccess()
						}

						select {
						case <-ctx.Done():
							return
						case req.Chan <- v:
							return
						}
					case <-ctx.Done():
						return
					}
				}()

				select {
				case outStream <- dirty.Request[P, R]{
					P:    req.P,
					Ctx:  req.Ctx,
					Chan: trackedChan,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func BreakerC[P, R any](
	engine BreakerEngine,
	breakerError error,
	chain dirty.ChainFunc[dirty.Request[P, R]],
) dirty.ChainFunc[dirty.Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan dirty.Request[P, R],
	) {
		chain(
			ctx,
			wg,
			Breaker[P, R](ctx, wg, engine, breakerError, inStream),
		)
	}
}

func Breakers[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	engine BreakerEngine,
	breakerError error,
	inStreams ...<-chan dirty.Request[P, R],
) (outStreams []<-chan dirty.Request[P, R]) {
	outStreams = make([]<-chan dirty.Request[P, R], len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = Breaker(ctx, wg, engine, breakerError, inStream)
	}

	return
}

func BreakersC[P, R any](
	engine BreakerEngine,
	breakerError error,
	chains dirty.ChainsFunc[dirty.Request[P, R]],
) dirty.ChainsFunc[dirty.Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan dirty.Request[P, R],
	) {
		chains(
			ctx,
			wg,
			Breakers[P, R](ctx, wg, engine, breakerError, inStreams...)...,
		)
	}
}
