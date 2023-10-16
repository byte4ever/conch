package dirty

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

func spreadMultiplex[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan T,
	outStreams []chan T,
) {
	defer wg.Done()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	outStreamsLen := len(outStreams)

	for t := range inStream {
		k := rnd.Intn(outStreamsLen)
		select {
		case <-ctx.Done():
			return
		case outStreams[k] <- t:
		}
	}
}

// FanInReducer multiplexes input streams into a lower number of output streams
// with balanced priority.
//
//	Output streams are closed when the context is done or all input streams are
//	closed.
func FanInReducer[T any](
	ctx context.Context,
	count int,
	inStreams ...<-chan T,
) []<-chan T {
	var wg sync.WaitGroup

	wg.Add(len(inStreams))

	multiplexedOutStreams := make([]chan T, count)
	for i := 0; i < count; i++ {
		multiplexedOutStreams[i] = make(chan T, len(inStreams))
	}

	for _, c := range inStreams {
		go spreadMultiplex(ctx, &wg, c, multiplexedOutStreams)
	}

	go func() {
		wg.Wait()

		for _, multiplexedOutStream := range multiplexedOutStreams {
			close(multiplexedOutStream)
		}
	}()

	out := make([]<-chan T, count)
	for i, stream := range multiplexedOutStreams {
		out[i] = stream
	}

	return out
}

func FanInReducerC[T any](
	count int,
	chains dirty.ChainsFunc[T],
) dirty.ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chains(ctx, wg, FanInReducer(ctx, count, inStream...)...)
	}
}
