package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

// GetProcessorFor return an async processor form the given function f.
func GetProcessorFor[From, To any](
	f func(context.Context, From) To,
) dirty.Processor[From, To] {
	return func(ctx context.Context, in <-chan From) <-chan To {
		outStream := make(chan To)

		go func() {
			defer close(outStream)

			for {
				select {
				case <-ctx.Done():
					return

				case v, more := <-in:
					if !more {
						return
					}

					select {
					case <-ctx.Done():
						return
					case outStream <- f(ctx, v):
					}
				}
			}
		}()

		return outStream
	}
}

// ProcessorPool launch the given processor concurrently and multiplexes the
// outputs in a single output stream.
func ProcessorPool[From, To any](
	ctx context.Context,
	concurrency int,
	processorFunc dirty.Processor[From, To],
	inStream <-chan From,
) (outputStream <-chan To) {
	if concurrency < 1 {
		panic("ProcessorPool: concurrency must be positive")
	}

	if concurrency == 1 {
		return processorFunc(ctx, inStream)
	}

	streamsToMerge := make([]<-chan To, concurrency)

	for i := 0; i < concurrency; i++ {
		streamsToMerge[i] = processorFunc(ctx, inStream)
	}

	return FanIn(ctx, streamsToMerge...)
}

func ProcessorPoolC[From, To any](
	concurrency int,
	processorFunc dirty.Processor[From, To],
	chain dirty.ChainFunc[To],
) dirty.ChainFunc[From] {
	return func(
		ctx context.Context, wg *sync.WaitGroup, inStream <-chan From,
	) {
		s := ProcessorPool(ctx, concurrency, processorFunc, inStream)
		chain(ctx, wg, s)
	}
}
