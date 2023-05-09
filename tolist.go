package conch

import (
	"context"
)

func ToList[T any](
	ctx context.Context,
	inStream <-chan T,
	expectedSize int,
) <-chan []T {
	outStream := make(chan []T)

	go func() {
		defer close(outStream)

		var outResult []T

		if expectedSize != 0 {
			outResult = make([]T, 0, expectedSize)
		}

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					select {
					case <-ctx.Done():
						return
					case outStream <- outResult:
						return
					}
				}

				outResult = append(outResult, v)
			}
		}
	}()

	return outStream
}
