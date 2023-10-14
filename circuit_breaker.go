package conch

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type (
	engineEvent  uint8
	CircuitState uint32
)

const (
	CircuitOpen     CircuitState = iota // OPEN
	CircuitClosed                       // CLOSED
	CircuitHalfOpen                     // HALF_OPEN
)

const ( //nolint:decorder,grouper // we want that
	engineEventSuccess engineEvent = iota // EVT_SUCCESS
	engineEventFailure                    // EVT_FAILURE
	engineEventRefresh                    // EVT_REFRESH
)

func circuitStateEngine(
	ctx context.Context,
	failuresToOpen int,
	successToClose int,
	halfOpenTimeout time.Duration,
) (
	func(context.Context),
	func(context.Context),
	func(context.Context),
	func() CircuitState,
) {
	comChan := make(chan engineEvent)
	var isDown atomic.Bool

	var (
		sharedState atomic.Uint32
	)

	sharedState.Store(uint32(CircuitClosed))

	go func() {
		defer close(comChan)
		defer isDown.Store(true)

		var (
			state              CircuitState
			consecutiveFailure int
			consecutiveSuccess int
			ts                 time.Time
		)

		state = CircuitClosed

		for {
			select {
			case <-ctx.Done():
				return
			case event := <-comChan:
				switch state {
				case CircuitOpen:
					if time.Now().After(ts) {
						state = CircuitHalfOpen
						consecutiveFailure = 0
						consecutiveSuccess = 0
						sharedState.Store(uint32(state))
					}

				case CircuitClosed:
					switch event {
					case engineEventSuccess:
						consecutiveFailure = 0
						consecutiveSuccess++

					case engineEventFailure:
						consecutiveSuccess = 0
						consecutiveFailure++
					}

					if consecutiveFailure == failuresToOpen {
						state = CircuitOpen
						consecutiveFailure = 0
						consecutiveSuccess = 0
						ts = time.Now().Add(halfOpenTimeout)

						sharedState.Store(uint32(state))
					}

				case CircuitHalfOpen:
					switch event {
					case engineEventSuccess:
						consecutiveFailure = 0
						consecutiveSuccess++

					case engineEventFailure:
						consecutiveSuccess = 0
						consecutiveFailure++
					}

					if consecutiveFailure > 0 {
						state = CircuitOpen
						consecutiveFailure = 0
						consecutiveSuccess = 0
						ts = time.Now().Add(halfOpenTimeout)

						sharedState.Store(uint32(state))

						continue
					}

					if consecutiveSuccess == successToClose {
						ts = time.Time{}
						consecutiveFailure = 0
						consecutiveSuccess = 0
						state = CircuitClosed

						sharedState.Store(uint32(state))
						continue
					}
				}
			}
		}
	}()

	return func(fCtx context.Context) {
			if isDown.Load() {
				return
			}
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventRefresh:
			}
		}, func(fCtx context.Context) {
			if isDown.Load() {
				return
			}
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventSuccess:
			}
		}, func(fCtx context.Context) {
			if isDown.Load() {
				return
			}
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventFailure:
			}
		}, func() CircuitState {
			return CircuitState(sharedState.Load())
		}
}

func Breaker[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan Request[P, R],
	nbFailureToOpen int,
	nbSuccessToClose int,
	halfOpenTimeout time.Duration,
	breakerError error,
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	go func() {
		defer close(outStream)
		refresh, success, failure, chState := circuitStateEngine(
			ctx,
			nbFailureToOpen,
			nbSuccessToClose,
			halfOpenTimeout,
		)

		chanPool := newValErrorChanPool[R](maxCapacity)

		for {
			select {
			case <-ctx.Done():
				return
			case req, more := <-inStream:
				if !more {
					return
				}

				switch chState() {
				case CircuitOpen:
					select {
					case <-ctx.Done():
						return
					case req.Chan <- ValErrorPair[R]{Err: breakerError}:
					}
					refresh(ctx)
				case CircuitClosed, CircuitHalfOpen:
					trackedChan := chanPool.get()

					wg.Add(1)

					go func() {
						defer wg.Done()
						defer chanPool.putBack(trackedChan)
						select {
						case v := <-trackedChan:
							if v.Err != nil {
								failure(ctx)
							} else {
								success(ctx)
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
		}
	}()

	return outStream
}

func BreakerC[P, R any](
	nbFailureToOpen int,
	nbSuccessToClose int,
	halfOpenTimeout time.Duration,
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
				nbFailureToOpen,
				nbSuccessToClose,
				halfOpenTimeout,
				breakerError,
			),
		)
	}
}
