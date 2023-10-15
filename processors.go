package conch

import (
	"context"
	"fmt"
	"sync"
)

func SingleProcessor[From, To any](
	ctx context.Context,
	wg *sync.WaitGroup,
	functor func(ctx context.Context, param From) (result To),
	inStream <-chan From,
) <-chan To {
	outStream := make(chan To)

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case v, more := <-inStream:
				if !more {
					return
				}

				select {
				case <-ctx.Done():
					return

				case outStream <- functor(ctx, v):
				}
			}
		}
	}()

	return outStream
}

func SingleProcessorC[From, To any](
	functor func(ctx context.Context, param From) (result To),
	chain ChainFunc[To],
) ChainFunc[From] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan From,
	) {
		chain(ctx, wg, SingleProcessor(ctx, wg, functor, inStream))
	}
}

func Processors[From, To any](
	ctx context.Context,
	wg *sync.WaitGroup,
	functor func(ctx context.Context, param From) (result To),
	inStreams ...<-chan From,
) []<-chan To {
	outStreams := make([]<-chan To, len(inStreams))

	for i, inStream := range inStreams {
		fmt.Println("processor ", i)

		outStreams[i] = SingleProcessor(ctx, wg, functor, inStream)
	}

	return outStreams
}

func ProcessorsC[From, To any](
	functor func(ctx context.Context, param From) (result To),
	chain ChainsFunc[To],
) ChainsFunc[From] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan From,
	) {
		chain(ctx, wg, Processors(ctx, wg, functor, inStreams...)...)
	}
}
