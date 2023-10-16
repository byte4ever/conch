package dirty

import (
	"context"
	"math/rand"
	"sync"

	"github.com/byte4ever/conch"
)

// Shuffle generates a stream by randomly picking elements from the
// input streams.
//
// Input streams can provide different number of elements.
// Input streams MUST differ (same stream cannot be present twice).
//
// Output streams is closed when all input streams have been exhausted and
// closed or the context is canceled.
//
//nolint:maintidx // dgas
func Shuffle[T any](
	ctx context.Context,
	rnd *rand.Rand,
	inStream ...<-chan T,
) <-chan T {
	outStream := make(chan T)

	if containsDuplicate(inStream) {
		panic("Shuffle: with duplicate input stream")
	}

	go func() {
		defer close(outStream)

		delStreamIdx := make([]int, 0, len(inStream))

		for {
			if len(inStream) == 0 {
				// not stream remains get out and close output stream
				return
			}

			rnd.Shuffle(
				len(inStream), func(i, j int) {
					inStream[i], inStream[j] = inStream[j], inStream[i]
				},
			)

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

func ShuffleC[T any](
	count int,
	rnd *rand.Rand,
	chain dirty.ChainFunc[T],
) []dirty.ChainFunc[T] {
	var (
		mainWg sync.WaitGroup
		iWg    *sync.WaitGroup
		iCtx   context.Context
	)

	r := make([]dirty.ChainFunc[T], count)
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
		chain(iCtx, iWg, Shuffle(iCtx, rnd, s...))
	}()

	return r
}
