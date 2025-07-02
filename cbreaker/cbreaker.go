package cbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"

	"github.com/byte4ever/conch/domain"
)

type (

	// stateType represents the internal state of the Engine.
	stateType int

	// Engine represents a circuit breaker that manages state transitions.
	// It tracks failure and success counts to determine circuit state shifts.
	Engine struct {
		clock clock.Clock
		config
		state              stateType
		consecutiveFailure int
		consecutiveSuccess int
		mtx                sync.RWMutex
	}
)

const (

	// ClosedState represents the initial or stable state of the circuit
	// breaker.
	ClosedState stateType = iota

	// OpenState indicates the engine is in an open state, halting further
	// operations.
	OpenState stateType = iota

	// HalfOpenState represents a state where the circuit partially allows
	// requests. It is used to test if the system has recovered after being in
	// OpenState.
	HalfOpenState stateType = iota
)

// NewEngine creates and initializes a new Engine instance with given options.
// It returns a pointer to the Engine and an error, if any occurs during setup.
func NewEngine(option ...Option) (*Engine, error) {
	return newEngine(clock.New(), option...) //nolint:wrapcheck // ok
}

// newEngine creates and configures a new Engine instance with provided options.
// Returns the Engine instance or an error if option application fails.
func newEngine(clock_ clock.Clock, option ...Option) (*Engine, error) {
	conf := defaultConfig

	for _, o := range option {
		if err := o.apply(&conf); err != nil {
			return nil, fmt.Errorf(
				"when creating new circuit breaker engine"+
					": %w",
				err,
			)
		}
	}

	return &Engine{
		config: conf,
		clock:  clock_,
		state:  ClosedState,
	}, nil
}

// IsOpen returns true if the engine's state is OpenState.
// This indicates that the engine is in an open state.
func (e *Engine) IsOpen() bool {
	e.mtx.RLock()
	defer e.mtx.RUnlock()

	return e.state == OpenState
}

// ReportFailure handles transition logic based on consecutive failures.
// If in ClosedState, increments failure count and transitions to OpenState
// when failures reach the configured threshold. In HalfOpenState, transitions
// to OpenState immediately. No action is taken if already in OpenState.
func (e *Engine) ReportFailure() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	switch e.state {
	case ClosedState:
		e.consecutiveFailure++

		if e.consecutiveFailure == *e.nbFailureToOpen {
			e.toOpenState()
		}
	case OpenState:
	case HalfOpenState:
		e.toOpenState()
	}
}

// ReportSuccess resets failures in ClosedState or increments successes in HalfOpenState.
// Transitions to ClosedState if the consecutive successes meet nbSuccessToClose.
func (e *Engine) ReportSuccess() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	switch e.state {
	case ClosedState:
		e.consecutiveFailure = 0

	case OpenState:

	case HalfOpenState:
		e.consecutiveSuccess++

		if e.consecutiveSuccess == *e.nbSuccessToClose {
			e.toCloseState()
		}
	}
}

// toCloseState transitions the engine state to ClosedState and resets counters.
func (e *Engine) toCloseState() {
	e.setStateReset(ClosedState)
}

// toOpenState transitions the engine state to OpenState and schedules a timeout.
// The timeout triggers a transition to HalfOpenState after halfOpenTimeout.
func (e *Engine) toOpenState() {
	e.setStateReset(OpenState)
	time.AfterFunc(*e.halfOpenTimeout, e.toHalfOpen)
}

// toHalfOpen transitions the Engine state to HalfOpenState.
func (e *Engine) toHalfOpen() {
	e.state = HalfOpenState
}

// setStateReset resets the engine's state and consecutive counters.
func (e *Engine) setStateReset(s stateType) {
	e.state = s
	e.consecutiveFailure = 0
	e.consecutiveSuccess = 0
}

var ErrCircuitBreakerIsOpen = errors.New("circuit breaker is open")

// WrapProcessorFunc wraps a ProcessorFunc with circuit breaker logic.
// It prevents processing if the engine is open and reports success/failure.
func WrapProcessorFunc[
	T domain.Processable[Param, Result],
	Param any,
	Result any](
	engine *Engine,
	f domain.ProcessorFunc[T, Param, Result],
) domain.ProcessorFunc[T, Param, Result] {
	return func(ctx context.Context, elem T) {
		if engine.IsOpen() {
			elem.SetError(ErrCircuitBreakerIsOpen)
			return
		}

		f(ctx, elem)

		if elem.GetError() != nil {
			engine.ReportFailure()
		} else {
			engine.ReportSuccess()
		}

		return
	}
}
