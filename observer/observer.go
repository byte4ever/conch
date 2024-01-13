package observer

import (
	"sync"
)

// ObservableEngine hold a engine compatible with Observable interface.
type ObservableEngine struct {
	mtx           sync.Mutex
	observers     map[Observer]struct{}
	observersChan map[Observer]chan interface{}
}

// NewObservable initialize the engine.
func NewObservable() *ObservableEngine {
	return &ObservableEngine{
		observers:     make(map[Observer]struct{}),
		observersChan: make(map[Observer]chan interface{}),
	}
}

// Register implement Observable interface.
func (m *ObservableEngine) Register(o Observer) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.observersChan, o)
	m.observers[o] = struct{}{}
	return nil
}

// RegisterChan implement Observable interface.
func (m *ObservableEngine) RegisterChan(o Observer, c chan interface{}) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.observers, o)
	m.observersChan[o] = c
	return nil
}

// Unregister implement Observable interface.
func (m *ObservableEngine) Unregister(o Observer) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.observers, o)
	delete(m.observersChan, o)
	return nil
}

// Notify implement Observable interface.
func (m *ObservableEngine) Notify(i interface{}) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for obs := range m.observers {
		go func(o Observer) {
			o.OnNotify(i)
		}(obs)
	}
	for _, c := range m.observersChan {
		go func(cc chan interface{}) {
			cc <- i
		}(c)
	}
}
