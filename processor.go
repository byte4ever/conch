// Package conch provides stream processing utilities for Go applications.
// This file implements data processing functionality for stream elements.
package conch

import (
	"context"
	"sync"

	"github.com/byte4ever/conch/domain"
)

// WrapClassicFuncToProcessorFunc adapts a classic function to work with the ProcessorFunc interface.
// It takes a function that accepts a parameter and returns a result with an error,
// and wraps it to work with Processable elements by setting the result or error appropriately.
func WrapClassicFuncToProcessorFunc[T domain.Processable[Param, Result], Param any, Result any](
	f func(
		ctx context.Context,
		param Param,
	) (Result, error)) domain.ProcessorFunc[T, Param, Result] {
	return func(ctx context.Context, elem T) {
		result, err := f(ctx, elem.GetParam())
		if err != nil {
			elem.SetError(err)
			return
		}
		elem.SetValue(result)
	}
}

// Processor creates a streaming processor that applies a processing function to each element.
// It reads from inStream, applies the processorFunc to modify each element in place,
// and forwards the modified elements to the output channel.
func Processor[T domain.Processable[Param, Result], Param any, Result any](
	ctx context.Context,
	wg *sync.WaitGroup,
	processorFunc domain.ProcessorFunc[T, Param, Result],
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	wg.Add(1)

	go func() {
		defer func() {
			close(outStream)
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case elem, more := <-inStream:
				if !more {
					return
				}

				processorFunc(ctx, elem)

				select {
				case <-ctx.Done():
					return
				case outStream <- elem:
				}
			}
		}
	}()

	return outStream
}

// ProcessorC creates a chainable processor function that can be composed with other stream operations.
// It wraps the Processor function to work with the ChainFunc interface for building processing pipelines.
func ProcessorC[T domain.Processable[Param, Result], Param any, Result any](
	processorFunc domain.ProcessorFunc[T, Param, Result],
	chain ChainFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chain(
			ctx,
			wg,
			Processor(
				ctx,
				wg,
				processorFunc,
				inStream,
			),
		)
	}
}

// Processors creates multiple parallel processors for processing multiple input streams.
// Each input stream is processed independently using the same processorFunc.
// Returns a slice of output channels corresponding to each input stream.
func Processors[T domain.Processable[Param, Result], Param any, Result any](
	ctx context.Context,
	wg *sync.WaitGroup,
	processorFunc domain.ProcessorFunc[T, Param, Result],
	inStreams ...<-chan T,
) (outStreams []<-chan T) {
	outStreams = make(
		[]<-chan T,
		0,
		len(inStreams),
	)

	for _, inStream := range inStreams {
		outStreams = append(
			outStreams,
			Processor(
				ctx,
				wg,
				processorFunc,
				inStream,
			),
		)
	}

	return
}

// ProcessorsC creates a chainable function for processing multiple streams.
// It wraps the Processors function to work with the ChainsFunc interface.
func ProcessorsC[T domain.Processable[Param, Result], Param any, Result any](
	processorFunc domain.ProcessorFunc[T, Param, Result],
	chains ChainsFunc[T],
) ChainsFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan T,
	) {
		chains(ctx, wg, Processors(
			ctx,
			wg,
			processorFunc,
			inStream...,
		)...)
	}
}
