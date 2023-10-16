package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

func SetIndex[T any, C dirty.Integer](
	ctx context.Context,
	inStream <-chan T,
) <-chan dirty.Indexed[C, T] {
	var idx C

	outStream := make(chan dirty.Indexed[C, T])

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
			case outStream <- dirty.Indexed[C, T]{
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

func SetIndexC[T any, C dirty.Integer](
	chain dirty.ChainFunc[dirty.Indexed[C, T]],
) dirty.ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(ctx, wg, SetIndex[T, C](ctx, inStream))
	}
}

func UnsetIndex[T any, C dirty.Integer](
	ctx context.Context,
	inStream <-chan dirty.Indexed[C, T],
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

func UnsetIndexC[T any, C dirty.Integer](
	chain dirty.ChainFunc[T],
) dirty.ChainFunc[dirty.Indexed[C, T]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan dirty.Indexed[C, T],
	) {
		chain(ctx, wg, UnsetIndex[T, C](ctx, inStream))
	}
}
