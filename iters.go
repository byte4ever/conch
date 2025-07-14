// Package conch provides stream processing utilities for Go applications.
// This file implements bidirectional conversion between Go iterators and channels.
package conch

import (
	"context"
	"iter"
	"sync"
)

// Pair represents a key-value pair used in iterator sequence conversions.
// It provides a convenient way to handle two-parameter iterators with channels.
type Pair[Key any, Value any] struct {
	Key   Key   // The key component of the pair.
	Value Value // The value component of the pair.
}

// ChanToIterSeq converts a channel to an iterator sequence.
// It reads values from the input channel and yields them through the iterator interface,
// respecting context cancellation and proper channel closure.
func ChanToIterSeq[T any](ctx context.Context, input <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-input:
				if !more {
					return
				}

				if !yield(v) {
					return
				}
			}
		}
	}
}

// IterSeqToChan converts an iterator sequence to a channel.
// It spawns a goroutine to iterate through the sequence and send values to the output channel,
// providing integration between Go's iterator interface and channel-based streams.
func IterSeqToChan[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	it iter.Seq[T],
) <-chan T {
	outputChan := make(chan T)

	wg.Add(1)

	go func() {
		defer func() {
			close(outputChan)
			wg.Done()
		}()

		for t := range it {
			select {
			case <-ctx.Done():
				return
			case outputChan <- t:
			}
		}
	}()

	return outputChan
}

// ChanToIterSeq2 converts a channel of Pair values to a two-parameter iterator sequence.
// It reads Pair values from the input channel and yields them as separate key-value parameters,
// respecting context cancellation and proper channel closure.
func ChanToIterSeq2[Key any, Value any](
	ctx context.Context,
	input <-chan *Pair[Key, Value],
) iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-input:
				if !more {
					return
				}

				if !yield(v.Key, v.Value) {
					return
				}
			}
		}
	}
}

// IterSeq2ToChan converts a two-parameter iterator sequence to a channel of Pair values.
// It spawns a goroutine to iterate through the sequence and send Pair values to the output channel,
// providing integration between Go's two-parameter iterator interface and channel-based streams.
func IterSeq2ToChan[Key any, Value any](
	ctx context.Context,
	wg *sync.WaitGroup,
	it iter.Seq2[Key, Value],
) <-chan *Pair[Key, Value] {
	outputChan := make(chan *Pair[Key, Value])

	wg.Add(1)

	go func() {
		defer func() {
			close(outputChan)
			wg.Done()
		}()

		for key, value := range it {
			select {
			case <-ctx.Done():
				return
			case outputChan <- &Pair[Key, Value]{
				Key:   key,
				Value: value,
			}:
			}
		}
	}()

	return outputChan
}
