package conch

import (
	"context"
	"sync"
)

// Injector creates an injector function to feed data into a channel stream.
// It manages synchronization and cleanup using context and a wait group.
// Accepts a context and returns an injection function and an output channel.
// The injection function sends items into the stream respecting context rules.
// The output channel emits the injected items, closing when the context ends.
func Injector[T any](
	ctx context.Context,
) (
	func(innerCtx context.Context,
		v T) error,
	<-chan T,
) {
	var wg sync.WaitGroup

	outStream := make(chan T)

	go func() {
		defer func() {
			wg.Wait()
			close(outStream)
		}()

		select {
		case <-ctx.Done():
			return
		}
	}()

	return func(
			innerCtx context.Context,
			v T,
		) error {
			wg.Add(1)
			defer wg.Done()

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-innerCtx.Done():
				return innerCtx.Err()
			default:
				select {
				case outStream <- v:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				case <-innerCtx.Done():
					return innerCtx.Err()
				}
			}
		},

		outStream
}

// InjectorC creates an injector function that feeds data into a chain of streams.
// The function uses a context and wait group for synchronization.
func InjectorC[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	chains ChainsFunc[T],
) func(context.Context, T) error {
	injector, ouStream := Injector[T](ctx)

	chains(
		ctx,
		wg,
		ouStream,
	)

	return injector
}
