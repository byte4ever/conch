package conch

import (
	"context"
	"sync"
)

func SetIndex[T any, C Integer](
	ctx context.Context,
	inStream <-chan T,
) <-chan Indexed[C, T] {
	var idx C

	outStream := make(chan Indexed[C, T])

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
			case outStream <- Indexed[C, T]{
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

func SetIndexC[T any, C Integer](
	chain ChainFunc[Indexed[C, T]],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, SetIndex[T, C](ctx, inStream))
	}
}

func UnsetIndex[T any, C Integer](
	ctx context.Context,
	inStream <-chan Indexed[C, T],
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

func UnsetIndexC[T any, C Integer](
	chain ChainFunc[T],
) ChainFunc[Indexed[C, T]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Indexed[C, T],
	) {
		chain(ctx, wg, UnsetIndex[T, C](ctx, inStream))
	}
}
