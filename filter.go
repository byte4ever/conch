package conch

import (
	"context"
)

func Filter[T any](
	ctx context.Context,
	inStream <-chan T,
	filter func(ctx context.Context, v T) bool,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case in, more := <-inStream:
				if !more {
					return
				}

				if filter(ctx, in) && ctx.Err() == nil {
					select {
					case outStream <- in:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return outStream
}
