package conch

import (
	"context"
)

func ToMap[In any, K comparable, V any](
	ctx context.Context,
	inStream <-chan In,
	mapper func(ctx context.Context, v In) (K, V),
	expectedCount int,
) <-chan map[K]V {
	outStream := make(chan map[K]V)

	go func() {
		var outMap map[K]V

		defer close(outStream)
		defer func() {
			outStream <- outMap
		}()

		if expectedCount == 0 {
			outMap = make(map[K]V)
		} else {
			outMap = make(map[K]V, expectedCount)
		}

		for {
			select {
			case <-ctx.Done():
				return

			case in, more := <-inStream:
				if !more {
					return
				}

				k, v := mapper(ctx, in)
				outMap[k] = v
			}
		}
	}()

	return outStream
}
