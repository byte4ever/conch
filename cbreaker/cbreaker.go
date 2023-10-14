package cbreaker

import (
	"sync"

	"github.com/andres-erbsen/clock"
)

type (
	stateType int

	Engine struct {
		config
		mtx                sync.Mutex
		state              stateType
		consecutiveFailure int
		consecutiveSuccess int
		clock              clock.Clock
	}
)

const (
	ClosedState   stateType = iota
	OpenState     stateType = iota
	HalfOpenState stateType = iota
)

func NewEngine(option ...Option) (*Engine, error) {
	return newEngine(clock.New(), option...)
}

func newEngine(clock clock.Clock, option ...Option) (*Engine, error) {
	conf := defaultConfig

	for _, o := range option {
		if err := o.apply(&conf); err != nil {
			return nil, err
		}
	}

	return &Engine{
		config: conf,
		clock:  clock,
		state:  ClosedState,
	}, nil
}

func (e *Engine) IsOpen() bool {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	return e.state == OpenState
}

func (e *Engine) ReportFailure() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	switch e.state {
	case ClosedState:
		e.consecutiveFailure++

		if e.consecutiveFailure == e.nbFailureToOpen {
			e.toOpenState()
		}
	case OpenState:
	case HalfOpenState:
		e.toOpenState()
	}
}

func (e *Engine) ReportSuccess() {
	e.mtx.Lock()
	defer e.mtx.Unlock()

	switch e.state {
	case ClosedState:
		e.consecutiveFailure = 0

	case OpenState:

	case HalfOpenState:
		e.consecutiveSuccess++

		if e.consecutiveSuccess == e.nbSuccessToClose {
			e.toCloseState()
		}
	}
}

func (e *Engine) toCloseState() {
	e.setStateReset(ClosedState)
}

func (e *Engine) toOpenState() {
	e.setStateReset(OpenState)

	go func() {
		e.clock.Sleep(e.halfOpenTimeout)
		e.toHalfOpen()
	}()
}

func (e *Engine) toHalfOpen() {
	e.state = HalfOpenState
}

func (e *Engine) setStateReset(s stateType) {
	e.state = s
	e.consecutiveFailure = 0
	e.consecutiveSuccess = 0
}
