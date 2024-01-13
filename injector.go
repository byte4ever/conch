package conch

import (
	"context"
	"sync"
)

// TODO :- lmartin 10/26/23 -: add closer to process.

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
