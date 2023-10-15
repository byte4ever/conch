package valve

import (
	"context"
)

type Dummy chan bool

func (d Dummy) Close() {
	select {
	case d <- false:
	default:
	}
}

func (d Dummy) Open() {
	select {
	case d <- true:
	default:
	}
}

func New() Dummy {
	return make(Dummy)
}

func (d Dummy) GetChannel(ctx context.Context) <-chan bool {
	outStream := make(chan bool)

	go func() {
		for {
			select {
			case v := <-d:
				select {
				case outStream <- v:
				case <-ctx.Done():
				}
			}
		}
	}()

	return outStream
}
