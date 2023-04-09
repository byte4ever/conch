package conch

import (
	"context"
)

func ToList[T any](
	ctx context.Context,
	inStream <-chan T,
	expectedSize int,
) <-chan []T {
	list := make(chan []T)

	go func() {
		var outResult []T

		defer close(list)
		defer func() {
			list <- outResult
		}()

		if expectedSize != 0 {
			outResult = make([]T, 0, expectedSize)
		}

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					return
				}

				outResult = append(outResult, v)
			}
		}
	}()

	return list
}
