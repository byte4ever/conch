package conch

import (
	"context"
)

// Buffer chains a buffered channel to inStream.
// This may help inStreams writer to be less frequently in a blocking state.
//
// Keep in mind that as soon as buffer is full it behave exactly as if it
// wasn't buffered at all.
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
