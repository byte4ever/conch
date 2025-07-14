// Package conch provides stream processing utilities for Go applications.
// This file implements iterator injection functionality for converting Go iterators to channels.
package conch

import (
	"context"
	"iter"
	"sync"
)

// iterInjectorCore handles the core logic of converting an iterator sequence to a channel.
// It iterates through the sequence and sends each value to the output channel,
// respecting context cancellation.
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

// IterInjector converts an iterator sequence to a channel for stream processing.
// It creates a buffered channel and spawns a goroutine to feed the iterator values
// into the channel, providing integration between Go's iterator interface and channel-based streams.
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

// IterInjectorC creates a chainable iterator injector that integrates with the ChainsFunc interface.
// It converts an iterator sequence to a channel and passes it to the chains function for further processing.
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
