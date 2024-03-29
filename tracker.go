package conch

import (
	"sync"
)

type (
	tracker[R any] struct {
		wg        sync.WaitGroup
		wgCollect sync.WaitGroup
		val       ValErrorPair[R]
	}
	trackerPool[R any] chan *tracker[R]
)

func newTracker[R any]() *tracker[R] {
	r := &tracker[R]{}

	//fmt.Println("++ inflight         newTracker")
	r.wg.Add(1)
	//fmt.Println("++ wgCollect  newTracker")
	r.wgCollect.Add(1)

	return r
}

func (t *tracker[R]) collectBy(pool trackerPool[R]) {
	t.wgCollect.Wait()
	pool.putBack(t)
}

func (t *tracker[R]) replicateValueIn(inChan chan<- ValErrorPair[R]) {
	//fmt.Println("++ wgCollect  replicateValueIn")
	t.wgCollect.Add(1)

	//fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>")
	t.wg.Wait()
	//fmt.Println("<<<<<<<<<<<<<<<<<<<<<<<<<<")
	inChan <- t.val

	// fmt.Println("-- wgCollect  replicateValueIn")
	t.wgCollect.Done()
}

func (t *tracker[R]) copyFromNewReturnStream(
	inChan chan<- ValErrorPair[R],
	outChan chan ValErrorPair[R],
	key Key,
	m *sync.Map,
	pool valErrorChanPool[R],
) {
	t.val = <-outChan
	pool.putBack(outChan)
	//fmt.Println("-- inflight         copyFromNewReturnStream")
	t.wg.Done()
	inChan <- t.val

	//fmt.Println("-- wgCollect  copyFromNewReturnStream")
	t.wgCollect.Done()
	m.Delete(key)
}

func newTrackerPool[R any](maxCapacity int) trackerPool[R] {
	return make(trackerPool[R], maxCapacity)
}

func (t trackerPool[R]) get() (nt *tracker[R]) {
	select {
	case nt = <-t:
	default:
		nt = newTracker[R]()
	}

	go nt.collectBy(t)

	return
}

func (t trackerPool[R]) putBack(v *tracker[R]) {
	var zeroR R

	v.val = ValErrorPair[R]{V: zeroR}

	//fmt.Println("++ inflight         putBack")
	v.wg.Add(1)
	//fmt.Println("++ wgCollect  putBack")
	v.wgCollect.Add(1)

	select {
	case t <- v:
	default: // trash it
	}
}
