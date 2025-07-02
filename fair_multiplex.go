package conch

import (
	"context"
	"fmt"
	"sync"
)

const (
	fairMultiplexNoInputStreamMsg = "conch.FairMultiplex: requires at least one input streams"
	fairMultiplexInvalidCountFmt  = "conch.FairMultiplex: count=%d: MUST be >0"
)

// fairMerge merges two input streams into a single output stream with
// balanced priority
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func fairMerge[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream1, inStream2 <-chan T,
) chan T {
	wg.Add(1)

	multiplexedOutStream := make(chan T)

	go func() {
		defer wg.Done()
		defer close(multiplexedOutStream)

		i1, i2 := inStream1, inStream2

		for {
			select {
			case <-ctx.Done():
				return
			case fromStream1, more := <-i1:
				if !more {
					i1 = nil

					if i2 == nil {
						return
					}

					continue
				}

				select {
				case multiplexedOutStream <- fromStream1:
				case <-ctx.Done():
					return
				}

			case fromStream2, more := <-i2:
				if !more {
					i2 = nil

					if i1 == nil {
						return
					}

					continue
				}

				select {
				case multiplexedOutStream <- fromStream2:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return multiplexedOutStream
}

// FairMultiplex multiplexes one or more input streams into one or more output streams
// with unbalanced priority.
//
// Input stream's priority is defined by its rank in parameter's list, from
// lowest to highest. Function will panic if no input stream are provided or
// count is less or equal to zero.
//
// The more back pressure is present on output stream (i.e. slow consumption)
// the more unfair priority effect will occur.
//
// If two input stream are provided, they will be multiplexed using exact same
// priority.
//
//	throughput(inputSteam#0) = throughput(inputSteam#1)
//
// For n > 1
//
//	throughput(inputSteam#n) = 2 * throughput(inputSteam#(n-1))
//
// If less back pressure is present, the more priority is balanced.
//
// Output streams is closed when all input stream are closed or context canceled.
//
// As for all *Multiplex function. It can be used for both fan-in and fan-out:
//   - For fan-in multiplex multiple input streams to a single output stream
//     (count=1).
//   - For fan-out multiplex a single input stream to multiple outputStreams
//     (count>=1).
func FairMultiplex[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	if len(inStreams) == 0 {
		panic(fairMultiplexNoInputStreamMsg)
	}

	if count <= 0 {
		panic(fmt.Sprintf(fairMultiplexInvalidCountFmt, count))
	}

	// bypass if no multiplex is required
	if len(inStreams) == 1 && count == 1 {
		return inStreams
	}

	multiplexedOutStreams := make([]chan T, count)

	// classic fan out
	if len(inStreams) == 1 {
		wg.Add(count)

		for i := 0; i < count; i++ {
			multiplexedOutStreams[i] = make(chan T)
			go func(cc chan T) {
				defer wg.Done()
				defer close(cc)
				for t := range inStreams[0] {
					select {
					case <-ctx.Done():
						return
					case cc <- t:
					}
				}
			}(multiplexedOutStreams[i])
		}

		out := make([]<-chan T, count)
		for i, stream := range multiplexedOutStreams {
			out[i] = stream
		}

		return out
	}

	// full multiplex multiple in streams to out streams
	for i := 0; i < count; i++ {
		outStream := fairMerge(ctx, wg, inStreams[0], inStreams[1])
		for _, inStream := range inStreams[2:] {
			outStream = fairMerge(ctx, wg, outStream, inStream)
		}

		multiplexedOutStreams[i] = outStream
	}

	out := make([]<-chan T, count)
	for i, stream := range multiplexedOutStreams {
		out[i] = stream
	}

	return out
}

// FairMultiplexC is the chainable version of FairMultiplex.
func FairMultiplexC[T any](
	count int,
	chains ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chains(ctx, wg, FairMultiplex(ctx, wg, count, inStream...)...)
	}
}
