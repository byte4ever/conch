package dirty

import "github.com/byte4ever/conch"

type valErrorChanPool[R any] chan chan dirty.ValErrorPair[R]

func newValErrorChanPool[R any](maxCapacity int) valErrorChanPool[R] {
	r := make(chan chan dirty.ValErrorPair[R], maxCapacity)
	return r
}

func (v valErrorChanPool[R]) get() chan dirty.ValErrorPair[R] {
	select {
	case c := <-v:
		return c
	default:
		return make(chan dirty.ValErrorPair[R], 1)
	}
}

func (v valErrorChanPool[R]) putBack(p chan dirty.ValErrorPair[R]) {
again:
	select {
	case <-p:
		goto again
	default:
		goto out
	}
out:
	select {
	case v <- p:
	default:
	}
}
