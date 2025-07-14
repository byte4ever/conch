// Package conch provides stream processing utilities for Go applications.
// This file implements data conversion between different types in streaming pipelines.
package conch

import (
	"context"
	"sync"
)

// ConverterFunc defines a function that converts data from one type to another.
// It takes a context for cancellation, input data of type InType, and writes
// the converted result to output of type OutType.
type ConverterFunc[InType any, OutType any] func(
	ctx context.Context,
	input *InType,
	output *OutType,
)

// Converter creates a streaming converter that transforms data from InType to OutType.
// It reads from inStream, applies the converterFunc transformation, and outputs to a new channel.
// The function uses object pools for memory efficiency and respects context cancellation.
// Returns a channel that emits converted OutType values.
func Converter[InType any, OutType any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inTypePool PoolReceiver[InType],
	outTypePool PoolProvider[OutType],
	converterFunc ConverterFunc[InType, OutType],
	inStream <-chan *InType,
) <-chan *OutType {
	outStream := make(chan *OutType)

	wg.Add(1)

	go func() {
		var inElem *InType
		var outElem *OutType
		var more bool

		defer func() {
			if inElem != nil {
				inTypePool.PutBack(inElem)
			}

			if outElem != nil {
				outTypePool.PutBack(outElem)
			}

			close(outStream)
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case inElem, more = <-inStream:
				if !more {
					return
				}

				outElem = outTypePool.Get()

				converterFunc(ctx, inElem, outElem)

				inTypePool.PutBack(inElem)
				inElem = nil

				select {
				case <-ctx.Done():
					return
				case outStream <- outElem:
					outElem = nil
				}
			}
		}
	}()

	return outStream
}

// ConverterC creates a chainable converter function that can be composed with other stream operations.
// It wraps the Converter function to work with the ChainFunc interface for building processing pipelines.
// Returns a ChainFunc that applies the conversion and passes results to the next chain function.
func ConverterC[InType, OutType any](
	inTypePool PoolReceiver[InType],
	outTypePool PoolProvider[OutType],
	converterFunc ConverterFunc[InType, OutType],
	chain ChainFunc[*OutType],
) ChainFunc[*InType] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan *InType,
	) {
		chain(
			ctx,
			wg,
			Converter(
				ctx,
				wg,
				inTypePool,
				outTypePool,
				converterFunc,
				inStream,
			),
		)
	}
}

// Converters creates multiple parallel converters for processing multiple input streams.
// Each input stream is converted independently using the same converterFunc.
// Returns a slice of output channels corresponding to each input stream.
func Converters[InType, OutType any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inTypePool PoolReceiver[InType],
	outTypePool PoolProvider[OutType],
	converterFunc ConverterFunc[InType, OutType],
	inStreams ...<-chan *InType,
) (outStreams []<-chan *OutType) {
	outStreams = make(
		[]<-chan *OutType,
		0,
		len(inStreams),
	)

	for _, inStream := range inStreams {
		outStreams = append(
			outStreams,
			Converter(
				ctx,
				wg,
				inTypePool,
				outTypePool,
				converterFunc,
				inStream,
			),
		)
	}

	return
}

// ConvertersC creates a chainable function for converting multiple streams.
// It wraps the Converters function to work with the ChainsFunc interface.
func ConvertersC[InType, OutType any](
	inTypePool PoolReceiver[InType],
	outTypePool PoolProvider[OutType],
	converterFunc ConverterFunc[InType, OutType],
	chains ChainsFunc[*OutType],
) ChainsFunc[*InType] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan *InType,
	) {
		chains(ctx, wg, Converters(
			ctx,
			wg,
			inTypePool,
			outTypePool,
			converterFunc,
			inStream...,
		)...)
	}
}
