// Package conch provides stream processing utilities for Go applications.
// This file implements sink functionality for consuming stream data.
package conch

import (
	"context"
	"sync"
)

// Sink creates a terminal stream processor that consumes data from a channel.
// It reads from inStream, applies the provided function to each element,
// and returns objects to the pool after processing. This is typically used
// as the final stage in a stream processing pipeline.
func Sink[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
	inStream <-chan *T,
) {
	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-inStream:
				if !more {
					return
				}
				f(ctx, v)
				poolReceiver.PutBack(v)
			}
		}
	}()
}

// SinkC creates a chainable sink function that can be composed with other
// stream operations. It wraps the Sink function to work with the ChainFunc
// interface for building processing pipelines.
func SinkC[T any](
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
) ChainFunc[*T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan *T) {
		Sink(
			ctx,
			wg,
			poolReceiver,
			f,
			inStream,
		)
	}
}

// Sinks creates multiple parallel sinks for consuming data from multiple input
// streams. Each input stream is processed independently using the same function
// and pool.
func Sinks[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
	inStreams ...<-chan *T,
) {
	for _, inStream := range inStreams {
		Sink(ctx, wg, poolReceiver, f, inStream)
	}
}

// SinksC creates a chainable function for consuming multiple streams.
// It wraps the Sinks function to work with the ChainsFunc interface.
func SinksC[T any](
	poolReceiver PoolReceiver[T],
	f func(context.Context, *T),
) ChainsFunc[*T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan *T,
	) {
		Sinks(
			ctx,
			wg,
			poolReceiver,
			f,
			inStream...,
		)
	}
}
