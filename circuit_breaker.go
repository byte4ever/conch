package conch

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type CircuitState uint32

const (
	CircuitUndefined CircuitState = iota // UNDEFINED
	CircuitOpen                          // OPEN
	CircuitClosed                        // CLOSED
	CircuitHalfOpen                      // HALF_OPEN
)

// ErrRequestCancelled represent an error where ....
var ErrRequestCancelled = errors.New("request cancelled")

type engineEvent uint8

const (
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
				// if ts.After(time.Now()) {
				// 	fmt.Println(
				// 		state, event, successCounter,
				// 		failureCounter, ts.Sub(time.Now()),
				// 	)
				// } else {
				// 	fmt.Println(state, event, successCounter, failureCounter)
				// }

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
					select {
					case req.ChResp <- ValErrorPair[R]{
						Err: ErrRequestCancelled,
					}:
						refresh(ctx)
					case <-ctx.Done():
						return
					}
				case CircuitClosed, CircuitHalfOpen:
					// Let request continue but intercept response to track
					// errors
					chResp := make(chan ValErrorPair[R])
					go func(c chan ValErrorPair[R]) {
						defer close(req.ChResp)
						select {
						case vep, more := <-c:
							// fmt.Println(">>> ", vep)
							if !more {
								return
							}

							if vep.Err == nil {
								success(ctx)
							} else {
								failure(ctx)
							}
							select {
							case req.ChResp <- vep:
							case <-ctx.Done():
								return
							}
						case <-ctx.Done():
							return
						}
					}(chResp)

					select {
					case outStream <- Request[P, R]{
						P:      req.P,
						ChResp: chResp,
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
