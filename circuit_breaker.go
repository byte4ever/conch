package conch

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type (
	engineEvent  uint8
	CircuitState uint32
)

const (
	CircuitUndefined CircuitState = iota // UNDEFINED
	CircuitOpen                          // OPEN
	CircuitClosed                        // CLOSED
	CircuitHalfOpen                      // HALF_OPEN
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
	<-chan CircuitState,
) {
	comChan := make(chan engineEvent)
	stateChan := make(chan CircuitState)

	var (
		sharedState atomic.Uint32
	)

	sharedState.Store(uint32(CircuitClosed))

	go func() {
		defer close(comChan)

		var (
			state          CircuitState
			failureCounter int
			successCounter int
			ts             time.Time
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
						failureCounter = 0
						successCounter = 0

						switch event {
						case engineEventSuccess:
							successCounter++
						case engineEventFailure:
							failureCounter++
						}

						sharedState.Store(uint32(state))
					}

				case CircuitClosed:
					switch event {
					case engineEventSuccess:
						successCounter++

						failureCounter = 0
					case engineEventFailure:
						failureCounter++

						successCounter = 0
					}

					if failureCounter == failuresToOpen {
						state = CircuitOpen
						failureCounter = 0
						successCounter = 0
						ts = time.Now().Add(halfOpenTimeout)

						sharedState.Store(uint32(state))
					}

				case CircuitHalfOpen:
					if failureCounter > 0 {

					}

					switch event {
					case engineEventSuccess:
						successCounter++

						failureCounter = 0

					case engineEventFailure:
						failureCounter++

						successCounter = 0
					}

					if failureCounter > 0 {
						state = CircuitOpen
						failureCounter = 0
						successCounter = 0
						ts = time.Now().Add(halfOpenTimeout)

						sharedState.Store(uint32(state))

						continue
					}

					if successCounter == successToClose {
						ts = time.Time{}
						failureCounter = 0
						successCounter = 0
						state = CircuitClosed

						sharedState.Store(uint32(state))

						continue
					}
				}
			}
		}
	}()

	go func() {
		defer close(stateChan)

		for {
			select {
			case <-ctx.Done():
				return
			case stateChan <- CircuitState(sharedState.Load()):
			}
		}
	}()

	return func(fCtx context.Context) {
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventRefresh:
			}
		}, func(fCtx context.Context) {
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventSuccess:
			}
		}, func(fCtx context.Context) {
			select {
			case <-fCtx.Done():
				return
			case <-ctx.Done():
				return
			case comChan <- engineEventFailure:
			}
		}, stateChan
}

func Breaker[P any, R any](
	ctx context.Context,
	inStream <-chan Request[P, R],
	nbFailureToOpen int,
	nbSuccessToClose int,
	halfOpenTimeout time.Duration,
	breakerError error,
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])

	go func() {
		defer close(outStream)
		defer func() {
			fmt.Println("breaker shutdown")
		}()

		refresh, success, failure, chState := circuitStateEngine(
			ctx,
			nbFailureToOpen,
			nbSuccessToClose,
			halfOpenTimeout,
		)

		for {
			select {
			case <-ctx.Done():
				return
			case req, more := <-inStream:
				if !more {
					return
				}

				switch <-chState {
				case CircuitOpen:
					// Drop new requests
					req.Return(ctx, ValErrorPair[R]{Err: breakerError})
					refresh(ctx)
				case CircuitClosed, CircuitHalfOpen:
					select {
					case outStream <- Request[P, R]{
						P: req.P,
						Return: func(
							ctx context.Context,
							v ValErrorPair[R],
						) {
							if v.Err != nil {
								failure(ctx)
							} else {
								success(ctx)
							}

							req.Return(ctx, v)
						},
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
