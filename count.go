package conch

import (
	"context"
)

func Count[T any](
	inStream <-chan T,
	ctx context.Context,
) (count <-chan int) {
	var c int

	iCount := make(chan int)

	go func() {
		defer close(iCount)

		for range inStream {
			c++
		}

		select {
		case iCount <- c:
		case <-ctx.Done():
			return
		}

		iCount <- c
	}()

	return iCount
}
