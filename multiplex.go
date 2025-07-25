package conch

import (
	"context"
	"fmt"
	"sync"
)

const (
	multiplexNoInputStreamMsg = "Multiplex: requires at least one input streams"
	multiplexInvalidCountFmt  = "Multiplex: count=%d: MUST be >0"
)

func multiplex[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan T,
	outStream chan T,
) {
	defer wg.Done()

	for t := range inStream {
		select {
		case <-ctx.Done():
			return
		case outStream <- t:
		}
	}
}

// Multiplex multiplexes one or more input streams into one or more output
// streams.
//
// Output streams is closed when all input stream are closed or context
// canceled.Function will panic if no input stream are provided or count is less
// or equal to zero.
//
// As for all *Multiplex function. It can be used for both fan-in and fan-out:
//   - For fan-in multiplex multiple input streams to a single output stream
//     (count=1).
//   - For fan-out multiplex a single input stream to multiple outputStreams
//     (count>=1).
func Multiplex[T any](
	ctx context.Context,
	count int,
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	if len(inStreams) == 0 {
		panic(multiplexNoInputStreamMsg)
	}

	if count <= 0 {
		panic(fmt.Sprintf(multiplexInvalidCountFmt, count))
	}

	// bypass if no multiplex is required
	if len(inStreams) == count {
		return inStreams
	}

	var wg sync.WaitGroup

	// optimize for classic fan in (i.e. single channel out).
	if count == 1 {
		outStream := make(chan T, len(inStreams))

		wg.Add(len(inStreams))

		for _, c := range inStreams {
			go multiplex(ctx, &wg, c, outStream)
		}

		go func() {
			wg.Wait()
			close(outStream)
		}()

		return []<-chan T{outStream}
	}

	// full multiplex multiple in streams to out streams
	wg.Add(len(inStreams) * count)

	multiplexedOutStreams := make([]chan T, count)
	for i := 0; i < count; i++ {
		multiplexedOutStreams[i] = make(chan T, len(inStreams))
	}

	for _, c := range inStreams {
		for _, outStream := range multiplexedOutStreams {
			go multiplex(ctx, &wg, c, outStream)
		}
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

// MultiplexC is the chainable version of Multiplex.
func MultiplexC[T any](
	count int,
	chains ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan T,
	) {
		chains(
			ctx,
			wg,
			Multiplex(ctx, count, inStream...)...,
		)
	}
}
