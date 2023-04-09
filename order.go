package conch

import (
	"context"
)

func SetIndex[T any](
	ctx context.Context,
	inStream <-chan T,
) <-chan Indexed[uint64, T] {
	var idx uint64

	outStream := make(chan Indexed[uint64, T])

	go func() {
		defer close(outStream)

	l1:
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-inStream:
			if !ok {
				return
			}
			select {
			case outStream <- Indexed[uint64, T]{
				Index:   idx,
				Payload: payload,
			}:
				idx++
				goto l1
			case <-ctx.Done():
				return
			}
		}
	}()

	return outStream
}

func UnsetIndex[Idx Ordered, T any](
	ctx context.Context,
	inStream <-chan Indexed[Idx, T],
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

	l1:
		select {
		case <-ctx.Done():
			return
		case indexed, ok := <-inStream:
			if !ok {
				return
			}
			select {
			case outStream <- indexed.Payload:
				goto l1
			case <-ctx.Done():
				return
			}
		}
	}()

	return outStream
}
