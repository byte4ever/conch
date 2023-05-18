package conch

import (
	"context"
	"sync"
)

const zeroInputStreamsPanicMsg = "zero input streams"

// FanIn multiplexes input streams into a single output stream with balanced
// priority.
//
//	Output stream is closed when the context is done or all input streams are
//	closed.
func FanIn[T any](
	ctx context.Context,
	inStreams ...<-chan T,
) <-chan T {
	var wg sync.WaitGroup

	wg.Add(len(inStreams))

	multiplexedOutStream := make(chan T, len(inStreams))

	for _, c := range inStreams {
		go multiplex(ctx, &wg, c, multiplexedOutStream)
	}

	go func() {
		wg.Wait()
		close(multiplexedOutStream)
	}()

	return multiplexedOutStream
}

func FanInC[T any](
	chain ChainFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chain(ctx, wg, FanIn(ctx, inStream...))
	}
}

// fairMerge merges two input streams into a single output stream with
// balanced priority
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func fairMerge[T any](
	inStream1, inStream2 <-chan T,
) chan T {
	multiplexedOutStream := make(chan T)

	go func() {
		defer close(multiplexedOutStream)

		i1, i2 := inStream1, inStream2

		for {
			if i1 == nil && i2 == nil {
				return
			}
			select {
			case t, more := <-i1:
				if !more {
					i1 = nil
					continue
				}
				multiplexedOutStream <- t
			case t, more := <-inStream2:
				if !more {
					i2 = nil
					continue
				}
				multiplexedOutStream <- t
			}
		}
	}()

	return multiplexedOutStream
}

// unfairMerge merges two input streams into a single output stream with
// unbalanced priority.
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func unfairMerge[T any](
	lowPrioInStream, highPrioInStream <-chan T,
) chan T {
	multiplexedOutStream := make(chan T)

	go func() {
		defer close(multiplexedOutStream)

		low, high := lowPrioInStream, highPrioInStream

		for {
			switch {
			case low == nil && high == nil:
				// both of them are closed so closeMe the output stream
				return

			case low == nil:
				// low prio stream is closed, so we use high prio stream only
				t, more := <-high
				if !more {
					high = nil
					continue
				}
				multiplexedOutStream <- t

			case high == nil:
				// high prio stream is closed, so we use low prio stream only
				t, more := <-low
				if !more {
					low = nil
					continue
				}
				multiplexedOutStream <- t
			default:
				select {
				// we try higher prio stream first
				case t, more := <-high:
					if !more {
						high = nil
						continue
					}
					multiplexedOutStream <- t
				default:
					// otherwise we try both of them
					select {
					case t, more := <-high:
						if !more {
							high = nil
							continue
						}
						multiplexedOutStream <- t
					case t, more := <-low:
						if !more {
							low = nil
							continue
						}
						multiplexedOutStream <- t
					}
				}
			}
		}
	}()

	return multiplexedOutStream
}

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

// FairFanIn merges multiple input streams into a single output stream with
// unbalanced priority.
//
//	Priority is defined by the input stream rank, from lowest to highest.
//	Function will panic if no input stream are provided.
//
//	The more back pressure is present on output stream (i.e. slow consumption)
//	the more unfair priority effect will occur.
//
//	If two input stream are provided, they will be merged into one output stream
//	with exact same priority.
//
//	If more are provided, stream in third position will have a twice the
//	throughput of the stream in second position.
//
//	throughput(inputSteam#n) = 2 * throughput(inputSteam#(n-1))
//
//	If less back pressure is present, the more priority is balanced.
//
//	Output stream is closed when all input stream are closed or context is done.
func FairFanIn[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	count := len(inStream)

	if count == 0 {
		panic(zeroInputStreamsPanicMsg)
	}

	if count == 1 {
		input := make(chan T)
		return ContextBreaker(ctx, input)
	}

	// decouple from input stream because unfairMerge doesn't use context for
	// clean up. We need to ensure the funnel will be destroyed by channel
	// closing cascading effect only.
	bridgeInStreams := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		bridgeInStreams[i] = ContextBreaker(ctx, inStream[i])
	}

	// creates fair merge funnel
	outStream := fairMerge(bridgeInStreams[0], bridgeInStreams[1])
	for _, inStream := range bridgeInStreams[2:] {
		outStream = fairMerge(inStream, outStream)
	}

	return outStream
}

func FairFanInC[T any](
	chain ChainFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chain(ctx, wg, FairFanIn(ctx, inStream...))
	}
}

// UnfairFanIn merges multiple input streams into a single output stream with
// unfair priority balance.
//
// Priority is defined by the input stream rank, from lowest to highest.
// Function will panic if no input stream are provided.
//
// The more back pressure is present on output stream (i.e. slow consumption)
// the more priority effect will occur.
//
// * inputSteam#n is got less priority than inputSteam#(n+1)
//
// If no back pressure is present, the more priority looks  balanced.
//
// Output stream is closed when all input stream are closed or context is done.
func UnfairFanIn[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	count := len(inStream)

	if count == 0 {
		panic(zeroInputStreamsPanicMsg)
	}

	if count == 1 {
		input := make(chan T)
		return ContextBreaker(ctx, input)
	}

	// decouple from input stream because unfairMerge doesn't use context for
	// clean up. We need to ensure the funnel will be destroyed by channel
	// closing cascading effect only.
	bridgeInStreams := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		bridgeInStreams[i] = ContextBreakerBuffer(
			ctx,
			count-1+1,
			inStream[i],
		)
	}

	// creates unfair merge funnel
	outStream := unfairMerge(bridgeInStreams[0], bridgeInStreams[1])
	for _, inStream := range bridgeInStreams[2:] {
		outStream = unfairMerge(outStream, inStream)
	}

	outStream2 := make(chan T)

	go func() {
		defer close(outStream2)

		defer func() {
			for range outStream {
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case e, more := <-outStream:
				if !more {
					return
				}
				select {
				case <-ctx.Done():
					return
				case outStream2 <- e:
				}
			}
		}
	}()

	return outStream2
}

func UnfairFanInC[T any](
	chain ChainFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream ...<-chan T,
	) {
		chain(ctx, wg, UnfairFanIn(ctx, inStream...))
	}
}
