package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

type ValveController func(ctx context.Context) <-chan bool

// Valve builds an opened valve between two streams.
func Valve[T any](
	ctx context.Context,
	isOpenChan <-chan bool,
	isOpen bool,
	inStream <-chan T,
) (
	outStream <-chan T,
) {
	var more bool

	iOutStream := make(chan T)
	vIsOpenChan := isOpenChan
	vInputStream := inStream

	if !isOpen {
		vInputStream = nil
	}

	go func() {
		defer close(iOutStream)

		for {
			select {
			case v, more := <-vInputStream:
				if !more {
					return
				}

				select {
				case iOutStream <- v:
				case <-ctx.Done():
					return
				}

			case isOpen, more = <-vIsOpenChan:
				if !more {
					vIsOpenChan = nil
					vInputStream = nil

					continue
				}

				if isOpen {
					vInputStream = inStream
				} else {
					vInputStream = nil
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return iOutStream
}

func ValveC[T any](
	isOpenChan <-chan bool,
	isOpen bool,
	chain dirty.ChainFunc[T],
) dirty.ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, Valve(ctx, isOpenChan, isOpen, inStream))
	}
}
