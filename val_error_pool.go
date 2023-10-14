package conch

type valErrorChanPool[R any] chan chan ValErrorPair[R]

func newValErrorChanPool[R any](maxCapacity int) valErrorChanPool[R] {
	r := make(chan chan ValErrorPair[R], maxCapacity)
	return r
}

func (v valErrorChanPool[R]) get() chan ValErrorPair[R] {
	select {
	case c := <-v:
		return c
	default:
		return make(chan ValErrorPair[R], 1)
	}
}

func (v valErrorChanPool[R]) putBack(p chan ValErrorPair[R]) {
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
