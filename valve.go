package conch

import (
	"context"
)

// Valve builds an opened valve between two streams.
func Valve[T any](
	ctx context.Context,
	inStream <-chan T,
	isOpen bool,
) (
	openIt func(),
	closeIt func(),
	outStream <-chan T,
) {
	iOutStream := make(chan T)
	blockingCtx, blockingCancel := context.WithCancel(ctx)
	isOpenChan := make(chan bool)

	go func() {
		defer close(iOutStream)
		defer blockingCancel()

	beg:
		if isOpen {
			select {
			case v, more := <-inStream:
				if !more {
					return
				}

			c1:
				select {
				case iOutStream <- v:
					goto beg
				case isOpen = <-isOpenChan:
					goto c1
				case <-ctx.Done():
					return
				}

			case isOpen = <-isOpenChan:
				goto beg
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case isOpen = <-isOpenChan:
		}

		goto beg
	}()

	return func() {
			select {
			case isOpenChan <- true:
			case <-blockingCtx.Done():
			}
		}, func() {
			select {
			case isOpenChan <- false:
			case <-blockingCtx.Done():
			}
		}, iOutStream
}
