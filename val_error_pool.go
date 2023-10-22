package conch

type valErrorChanPool[R any] chan chan ValErrorPair[R]

func newValErrorChanPool[R any](maxCapacity int) valErrorChanPool[R] {
	return make(
		chan chan ValErrorPair[R],
		maxCapacity,
	)
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
	go func() {
		for {
			select {
			// clean chan p content by draining it
			case <-p:
			default:
				select {
				// try to put it back in pool
				case v <- p:
					return
				default:
					return
				}
			}
		}
	}()
}
