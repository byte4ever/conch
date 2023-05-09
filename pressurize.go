package conch

import (
	"context"
	"time"
)

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
