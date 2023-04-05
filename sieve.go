package conch

import (
	"context"
)

type IndexGetter[O any] func(O) int

func Sieve[T any](
	ctx context.Context,
	choice IndexGetter[T],
	count int,
	inStream <-chan T,
) []<-chan T {
	outStreams := make([]chan T, count)
	// create output streams
	for i := 0; i < count; i++ {
		outStreams[i] = make(chan T, count)
	}

	go func() {
		defer func() {
			for i := 0; i < count; i++ {
				close(outStreams[i])
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					return
				}

				select {
				case outStreams[choice(v)] <- v:

				case <-ctx.Done():
					return
				}
			}
		}

	}()

	outStreamsDir := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		outStreamsDir[i] = outStreams[i]
	}
	return outStreamsDir
}
