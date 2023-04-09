package conch

import (
	"context"
)

func Consume[T any](
	ctx context.Context,
	inStream <-chan T,
	consumer func(ctx context.Context, v T),
) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-inStream:
			if !ok {
				return
			}

			consumer(ctx, v)
		}
	}
}
