package conch

import (
	"context"
)

func Buffer[T any](
	ctx context.Context,
	inStream <-chan T,
	size int,
) <-chan T {
	outStream := make(chan T, size)

	go func() {
		defer close(outStream)

		for v := range inStream {
			select {
			case <-ctx.Done():
				return
			case outStream <- v:
			}
		}
	}()

	return outStream
}
