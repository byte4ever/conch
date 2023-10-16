package dirty

import (
	"context"
)

type Pair[A, B any] struct {
	A, B any
}

func Zip[A, B any](
	ctx context.Context,
	inStreamA <-chan A,
	inStreamB <-chan B,
) <-chan Pair[A, B] {
	outStream := make(chan Pair[A, B])

	go func() {
		defer close(outStream)

		for {
			var (
				a A
				b B
			)

			ca, cb := inStreamA, inStreamB

			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
					return
				case a = <-ca:
					ca = nil
				case b = <-cb:
					cb = nil
				}
			}

			select {
			case outStream <- Pair[A, B]{a, b}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outStream
}
