package conch

import (
	"context"
	"fmt"
	"sync"
)

type KK[P Hashable, R any] struct {
	m   map[Key]*Sub[R]
	mtx sync.Mutex
}

func NewKK[P Hashable, R any]() *KK[P, R] {
	return &KK[P, R]{
		m: make(map[Key]*Sub[R]),
	}
}

func (t *KK[P, R]) Subscribe(
	ctx context.Context,
	pool subPool[R],
	param P,
	outChan chan<- ValErrorPair[R],
) chan ValErrorPair[R] {
	sub := pool.get()
	sub.inflight.Add(1)
	key := param.Hash()

	t.mtx.Lock()
	actualSub, found := t.m[key]
	defer t.mtx.Unlock()

	if found {
		sub.inflight.Done()
		pool.putBack(sub)
		actualSub.inflight.Add(1)
		go func() {
			actualSub.cond.L.Lock()
			for !actualSub.available {
				actualSub.cond.Wait()
			}
			outChan <- actualSub.valueErr
			actualSub.inflight.Done()
			actualSub.cond.L.Unlock()
		}()
		return nil
	}

	t.m[key] = sub

	var all sync.WaitGroup
	all.Add(3)
	go func() {
		all.Done()
		sub.inflight.Wait()
		pool.putBack(sub)
	}()

	go func(k Key) {
		all.Done()
		sub.cond.L.Lock()
		sub.valueErr = <-sub.inStream
		sub.available = true
		t.mtx.Lock()
		delete(t.m, k)
		t.mtx.Unlock()
		sub.cond.Broadcast()
		sub.cond.L.Unlock()
	}(key)

	go func() {
		all.Done()
		sub.cond.L.Lock()
		for !sub.available {
			sub.cond.Wait()
		}
		outChan <- sub.valueErr
		sub.inflight.Done()
		sub.cond.L.Unlock()
	}()

	all.Wait()
	return sub.inStream
}

type Sub[R any] struct {
	inStream  chan ValErrorPair[R]
	inflight  *sync.WaitGroup
	valueErr  ValErrorPair[R]
	available bool
	cond      *sync.Cond
}
type subPool[R any] chan *Sub[R]

func (t subPool[R]) get() (nt *Sub[R]) {
	select {
	case nt = <-t:
	default:
		nt = &Sub[R]{
			inStream: make(chan ValErrorPair[R]),
			cond:     sync.NewCond(&sync.Mutex{}),
			inflight: &sync.WaitGroup{},
		}
	}

	return
}

func (t subPool[R]) putBack(v *Sub[R]) {
	v.available = false

	select {
	case t <- v:
	default: // trash it
		v.inStream = nil
		v.inflight = nil
		fmt.Println("trashing")
	}
}

func newSubPool[R any](maxCapacity int) subPool[R] {
	return make(subPool[R], maxCapacity)
}
