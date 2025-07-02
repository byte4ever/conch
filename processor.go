package conch

import (
	"context"
	"sync"

	"github.com/byte4ever/conch/domain"
)

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
