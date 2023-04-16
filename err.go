package conch

import (
	"context"
)

type ValErrorPair[V any] struct {
	V   V
	Err error
}

func Separate[V any](
	ctx context.Context,
	inSTream <-chan ValErrorPair[V],
) (<-chan V, <-chan error) {
	outStream := make(chan V)
	outStreamErr := make(chan error)

	go func() {
		defer close(outStream)
		defer close(outStreamErr)

		for v := range inSTream {
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
