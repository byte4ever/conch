package conch

import (
	"context"
	"iter"
	"sync"
)

func iterInjectorCore[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	seq iter.Seq[T],
	outputStream chan T,
) {
	defer func() {
		close(outputStream)
		wg.Done()
	}()

	for v := range seq {
		select {
		case <-ctx.Done():
			return
		case outputStream <- v:
		}
	}
}

func IterInjector[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	seq iter.Seq[T],
) <-chan T {
	wg.Add(1)

	output := make(chan T, 100)

	go iterInjectorCore(
		ctx,
		wg,
		seq,
		output,
	)
	return output
}

func IterInjectorC[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	seq iter.Seq[T],
	chains ChainsFunc[T],
) {
	chains(
		ctx,
		wg,
		IterInjector[T](ctx, wg, seq),
	)
}
