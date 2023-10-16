package dirty

import (
	"context"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

// Pressurize creates some initial pressure by creating a delay before
// to open the stream and copy it to output
func Pressurize[T any](
	ctx context.Context,
	inStream <-chan T,
	delay time.Duration,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return

		case <-timer.C:
		l2:
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					return
				}

				select {

				case outStream <- v:
					goto l2

				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func PressurizeC[T any](
	delay time.Duration,
	chain dirty.ChainFunc[T],
) dirty.ChainFunc[T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan T) {
		s := Pressurize(ctx, inStream, delay)
		chain(ctx, wg, s)
	}
}
