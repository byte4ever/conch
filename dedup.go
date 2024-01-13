package conch

import (
	"context"
	"sync"
)

const (
	maxCapacity = 100
)

func Deduplicator[P Hashable, R any](
	ctx context.Context,
	_ *sync.WaitGroup,
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])
	go func() {
		defer close(outStream)

		m := NewKK[P, R]()
		subPool := newSubPool[R](maxCapacity)

	again:
		select {
		case <-ctx.Done():
			return
		case req, more := <-inStream:
			if !more {
				return
			}

			if dedupReturn := m.Subscribe(
				ctx,
				subPool,
				req.P,
				req.Chan,
			); dedupReturn != nil {
				select {
				case <-ctx.Done():
					return
				case outStream <- Request[P, R]{
					P:    req.P,
					Chan: dedupReturn,
				}:
					// pushed
				}
			}

		}

		goto again
	}()

	return outStream
}

func DeduplicatorC[P Hashable, R any](
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, Deduplicator(ctx, wg, inStream))
	}
}

func Deduplicators[P Hashable, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStreams ...<-chan Request[P, R],
) (outStreams []<-chan Request[P, R]) {
	outStreams = make([]<-chan Request[P, R], len(inStreams))

	for i, inStream := range inStreams {
		outStreams[i] = Deduplicator(ctx, wg, inStream)
	}

	return
}

func DeduplicatorsC[P Hashable, R any](
	chains ChainsFunc[Request[P, R]],
) ChainsFunc[Request[P, R]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStreams ...<-chan Request[P, R],
	) {
		chains(ctx, wg, Deduplicators(ctx, wg, inStreams...)...)
	}
}
