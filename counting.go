package conch

import (
	"context"
)

func Counting[T any](
	ctx context.Context,
	inStream <-chan T,
) <-chan uint64 {
	outStream := make(chan uint64)

	go func() {
		var cnt uint64

		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case _, more := <-inStream:
				if !more {
					return
				}

				select {
				case outStream <- cnt:
					cnt++
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}
