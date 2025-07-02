package conch

import (
	"context"
	"sync"
)

type DecoratorFunc[Item any] func(ctx context.Context, p *Item)

// decoratorCore coordinates item processing through a decorator function.
// ctx manages cancellation of the processing operation as needed.
// wg tracks lifecycle and ensures completion of the goroutine.
// decoratorFunc is applied to each item from the input channel inStream.
// inStream provides the incoming items to be processed and decorated.
// outStream receives the processed and decorated items for further usage.
func decoratorCore[Item any](
	ctx context.Context,
	wg *sync.WaitGroup,
	decoratorFunc DecoratorFunc[Item],
	inStream <-chan *Item,
	outStream chan *Item,
) {
	defer func() {
		close(outStream)
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case item, more := <-inStream:
			if !more {
				return
			}

			decoratorFunc(ctx, item)

			select {
			case <-ctx.Done():
				return

			case outStream <- item:
			}
		}
	}
}

// Decorator decorates items from an input channel using a provided function.
// It processes items concurrently, controlled by a context and wait group.
// ctx manages cancellation of the processing operation.
// wg tracks the lifetime of the processing goroutine.
// decoratorFunc is applied to each item received from the input channel.
// inStream is the input channel providing items to be processed.
// Returns an output channel where decorated items are sent.
func Decorator[Item any](
	ctx context.Context,
	wg *sync.WaitGroup,
	decoratorFunc DecoratorFunc[Item],
	inStream <-chan *Item,
) <-chan *Item {
	wg.Add(1)

	outStream := make(chan *Item)
	go decoratorCore(
		ctx,
		wg,
		decoratorFunc,
		inStream,
		outStream,
	)

	return outStream
}

// DecoratorC applies a decorator function to items and chains to another function.
// decoratorFunc is the function used to process each incoming item.
// chain is the next processing function that consumes decorated items.
func DecoratorC[Item any](
	decoratorFunc DecoratorFunc[Item],
	chain ChainFunc[*Item],
) ChainFunc[*Item] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan *Item,
	) {
		chain(
			ctx,
			wg,
			Decorator(
				ctx,
				wg,
				decoratorFunc,
				inStream,
			),
		)
	}
}

// Decorators applies a decoratorFunc to items from multiple input channels.
// ctx manages the cancellation of the entire operation.
// wg tracks the lifetime of all processing goroutines.
// decoratorFunc is applied to each item received from input channels.
// inStreams represents multiple input channels providing items to process.
// Returns a slice of output channels containing the decorated items.
func Decorators[Item any](
	ctx context.Context,
	wg *sync.WaitGroup,
	decoratorFunc DecoratorFunc[Item],
	inStreams ...<-chan *Item,
) (outStreams []<-chan *Item) {
	outStreams = make([]<-chan *Item, 0, len(inStreams))

	for _, inStream := range inStreams {
		outStreams = append(
			outStreams,
			Decorator(
				ctx,
				wg,
				decoratorFunc,
				inStream,
			),
		)
	}

	return
}

// DecoratorsC applies a decorator function to items processed by a chain.
// It wraps ChainsFunc with additional decoration logic.
// ctx manages cancellation of the processing operation.
// decoratorFunc modifies each item using custom logic.
// chains represent the sequence of operations to execute.
// Returns a ChainsFunc with the decorator applied.
func DecoratorsC[Item any](
	decoratorFunc DecoratorFunc[Item],
	chains ChainsFunc[*Item],
) ChainsFunc[*Item] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan *Item,
	) {
		chains(ctx, wg, Decorators(
			ctx,
			wg,
			decoratorFunc,
			inStream...,
		)...)
	}
}
