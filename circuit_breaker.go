package conch

import (
	"context"
	"sync"
)

type BreakerEngine interface {
	IsOpen() bool
	ReportFailure()
	ReportSuccess()
}

func Breaker[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan Request[P, R],
	engine BreakerEngine,
	breakerError error,
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	go func() {
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
					case req.Chan <- ValErrorPair[R]{Err: breakerError}:
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
				case outStream <- Request[P, R]{
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
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(
			ctx,
			wg,
			Breaker[P, R](
				ctx,
				wg,
				inStream,
				engine,
				breakerError,
			),
		)
	}
}
