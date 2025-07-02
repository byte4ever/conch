package conch

import (
	"context"
	"sync"
)

type ConverterFunc[InType any, OutType any] func(
	ctx context.Context,
	input *InType,
	output *OutType,
)

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
