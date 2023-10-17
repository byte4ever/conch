package conch

import (
	"context"
	"fmt"
	"sync"
)

const (
	unfairMultiplexNoInputStreamMsg = "conch.UnfairMultiplex: requires at least one input streams"
	unfairMultiplexInvalidCountFmt  = "conch.UnfairMultiplex: count=%d: MUST be >0"
)

// unfairMerge merges two input streams into a single output stream with
// unbalanced priority.
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func unfairMerge[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	lowPrioInStream, highPrioInStream <-chan T,
) chan T {
	multiplexedOutStream := make(chan T)

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(multiplexedOutStream)

		low, high := lowPrioInStream, highPrioInStream

		for {
			switch {
			case low == nil && high == nil:
				// both of them are closed so closeMe the output stream
				return

			case low == nil:
				// low prio stream is closed, so we use high prio stream only
				select {
				case t, more := <-high:
					if !more {
						high = nil
						continue
					}
					select {
					case multiplexedOutStream <- t:
					case <-ctx.Done():
						return
					}
				}

			case high == nil:
				// high prio stream is closed, so we use low prio stream only
				select {
				case t, more := <-low:
					if !more {
						low = nil
						continue
					}
					select {
					case multiplexedOutStream <- t:
					case <-ctx.Done():
						return
					}
				}
			default:
				select {
				// we try higher prio stream first
				case t, more := <-high:
					if !more {
						high = nil
						continue
					}
					select {
					case multiplexedOutStream <- t:
					case <-ctx.Done():
						return
					}
				default:
					// otherwise we try both of them
					select {
					case <-ctx.Done():
						return
					case t, more := <-high:
						if !more {
							high = nil
							continue
						}
						select {
						case multiplexedOutStream <- t:
						case <-ctx.Done():
							return
						}
					case t, more := <-low:
						if !more {
							low = nil
							continue
						}
						select {
						case multiplexedOutStream <- t:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return multiplexedOutStream
}

// UnfairMultiplex multiplexes one or more input streams into one or more output
// streams with unbalanced priority.
//
// Input stream's priority is defined by its rank in parameter's list, from
// lowest to highest. Function will panic if no input stream are provided or
// count is less or equal to zero.
// The more back pressure is present on output stream (i.e. slow consumption)
// the more priority effect will occur.
//
//	priority(inputSteam#n) <  inputSteam#(n+1)
//
// If no back pressure is present, the more priority looks balanced. Output
// streams is closed when all input stream are closed or context canceled.
// Function will panic if no input stream are provided or count is less or equal
// to zero.
//
// As for all *Multiplex function. It can be used for both fan-in and fan-out:
//   - For fan-in multiplex multiple input streams to a single output stream
//     (count=1).
//   - For fan-out multiplex a single input stream to multiple outputStreams
//     (count>=1).
func UnfairMultiplex[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	if len(inStreams) == 0 {
		panic(unfairMultiplexNoInputStreamMsg)
	}

	if count <= 0 {
		panic(fmt.Sprintf(unfairMultiplexInvalidCountFmt, count))
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
		outStream := unfairMerge(ctx, wg, inStreams[0], inStreams[1])
		for _, inStream := range inStreams[2:] {
			outStream = unfairMerge(ctx, wg, outStream, inStream)
		}

		multiplexedOutStreams[i] = outStream
	}

	out := make([]<-chan T, count)
	for i, stream := range multiplexedOutStreams {
		out[i] = stream
	}

	return out
}

// UnfairMultiplexC is the chainable version of UnfairMultiplex.
func UnfairMultiplexC[T any](
	count int,
	chains ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chains(ctx, wg, UnfairMultiplex(ctx, wg, count, inStream...)...)
	}
}
