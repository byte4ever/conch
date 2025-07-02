package conch

import (
	"context"
	"iter"
	"sync"
)

type Pair[Key any, Value any] struct {
	Key   Key
	Value Value
}

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
