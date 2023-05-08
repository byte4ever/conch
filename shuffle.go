package conch

import (
	"context"
	"math/rand"
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
		panic("distribute with duplicate input stream")
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
