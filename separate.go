package conch

import (
	"context"
	"sync"
)

func Separate[V any](
	ctx context.Context,
	inStream <-chan ValErrorPair[V],
) (<-chan V, <-chan error) {
	outStream := make(chan V)
	outStreamErr := make(chan error)

	go func() {
		defer close(outStream)
		defer close(outStreamErr)

		for v := range inStream {
			if v.Err != nil {
				select {
				case outStreamErr <- v.Err:
				case <-ctx.Done():
					return
				}

				continue
			}

			select {
			case outStream <- v.V:
			case <-ctx.Done():
				return
			}

			continue
		}
	}()

	return outStream, outStreamErr
}

func SeparateC[R any](
	result ChainFunc[R],
	err ChainFunc[error],
) ChainFunc[ValErrorPair[R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan ValErrorPair[R],
	) {
		r, e := Separate(ctx, inStream)
		err(ctx, wg, e)
		result(ctx, wg, r)
	}
}
