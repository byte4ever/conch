package conch

import (
	"context"
	"sync"
)

// Distribute generates a stream by sequentially picking elements from the
// input streams.
//
// Input streams can provide different number of elements.
// Input streams MUST differ (cannot be te same stream).
//
// Output streams is closed when all input streams have been exhausted and
// closed or the context is canceled.
func Distribute[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	outStream := make(chan T)

	if containsDuplicate(inStream) {
		panic("Distribute: with duplicate input stream")
	}

	go func() {
		defer close(outStream)

		delStreamIdx := make([]int, 0, len(inStream))

		for {
			if len(inStream) == 0 {
				// not stream remains get out and close output stream
				return
			}

			for idx, in := range inStream {
				select {
				case <-ctx.Done():
					return
				case e, more := <-in:
					if !more {
						// mark stream for deletion
						delStreamIdx = append(delStreamIdx, idx)
						continue
					}

					select {
					case <-ctx.Done():
						return
					case outStream <- e:
					}
				}
			}

			// apply stream deletions
			if len(delStreamIdx) != 0 {
				for i := len(delStreamIdx) - 1; i >= 0; i-- {
					inStream = append(
						inStream[:delStreamIdx[i]],
						inStream[delStreamIdx[i]+1:]...,
					)
				}

				// reset stream deletions indexes
				delStreamIdx = delStreamIdx[:0]
			}
		}
	}()

	return outStream
}

type ChainOp[T any] func(
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan T,
	outStream *<-chan T,
)

func OutC[T any](outputStream chan<- T) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case e, more := <-inStream:
					if !more {
						return
					}

					select {
					case <-ctx.Done():
						return
					case outputStream <- e:
					}
				}
			}
		}()
	}
}

func OpC[T any](insert ChainOp[T], others ...ChainFunc[T]) []ChainFunc[T] {
	lo := len(others)
	out := make([]ChainFunc[T], lo)

	for i := 0; i < lo; i++ {
		idx := i
		out[i] = func(
			ctx context.Context, wg *sync.WaitGroup, inStream <-chan T,
		) {
			var outStream <-chan T

			insert(ctx, wg, inStream, &outStream)
			others[idx](ctx, wg, outStream)
		}
	}

	return out
}

func DistributeC[T any](
	count int,
	chain ChainFunc[T],
) []ChainFunc[T] {
	var (
		mainWg sync.WaitGroup
		iWg    *sync.WaitGroup
		iCtx   context.Context
	)

	r := make([]ChainFunc[T], count)
	s := make([]<-chan T, count)

	mainWg.Add(count)

	for i := 0; i < count; i++ {
		j := i
		r[i] = func(
			ctx context.Context, wg *sync.WaitGroup, inStream <-chan T,
		) {
			defer mainWg.Done()

			if j == 0 {
				// capture context and wait group
				iCtx = ctx
				iWg = wg
			}

			s[j] = inStream
		}
	}

	go func() {
		mainWg.Wait()
		chain(iCtx, iWg, Distribute(iCtx, s...))
	}()

	return r
}
