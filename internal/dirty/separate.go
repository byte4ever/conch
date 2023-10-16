package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

func Separate[V any](
	ctx context.Context,
	inStream <-chan dirty.ValErrorPair[V],
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
	result dirty.ChainFunc[R],
	err dirty.ChainFunc[error],
) dirty.ChainFunc[dirty.ValErrorPair[R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan dirty.ValErrorPair[R],
	) {
		r, e := Separate(ctx, inStream)
		err(ctx, wg, e)
		result(ctx, wg, r)
	}
}
