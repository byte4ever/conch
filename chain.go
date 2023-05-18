package conch

import (
	"context"
	"sync"
)

// Chain performs streams concatenation.
// It generates a new stream that will produce all the elements from the
// streams consumed in the order they are provided.
//
// For example if
//
//	s1 provides elements s1e1, s1e2, s1e3 then close
//	s2 provides elements s2e1, s2e2 then close
//
// the resulting stream will produce elements
//
//	s1e1, s1e2, s1e3, s2e1, s2e2 then close
//
// pay attention all input streams must be closed and in the given example s2
// elements will be streamed to output only if s1 is closed.
//
// The returned stream will be closed when all input streams are closed or
// context is canceled.
func Chain[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	outStream := make(chan T)

	if containsDuplicate(inStream) {
		panic("chain with duplicate input")
	}

	go func() {
		defer close(outStream)

		for _, in := range inStream {
			for {
				select {
				case <-ctx.Done():
					return
				case e, more := <-in:
					if !more {
						goto out
					}
					select {
					case <-ctx.Done():
						return
					case outStream <- e:
					}
				}
			}
		out:
		}
	}()

	return outStream
}

func ChainC[T any](
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
		chain(iCtx, iWg, Chain(iCtx, s...))
	}()

	return r
}
