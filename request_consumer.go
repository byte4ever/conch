package conch

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNilRequestProcessingFunc = errors.New("nil request processing func")
)

func interceptPanic82763487263[P, R any](
	f RequestFunc[P, R],
) RequestFunc[P, R] {
	const (
		MaxCallerStackSize = 128
		StringBufferSize   = 2048
	)

	return func(ctx context.Context, p P) (res R, err error) {
		defer func() {
			if r := recover(); r != nil {
				var sb strings.Builder

				sb.Grow(StringBufferSize)

				stk := make([]uintptr, MaxCallerStackSize)
				n := runtime.Callers(4, stk)
				stk = stk[:n]

				for _, pc := range stk {
					f := runtime.FuncForPC(pc)
					file, l := f.FileLine(pc)
					name := f.Name()

					if strings.Contains(
						name,
						".interceptPanic82763487263[...]",
					) {
						continue
					}

					sb.WriteString(name)
					sb.WriteString("\n\t")
					sb.WriteString(file)
					sb.WriteRune(':')
					sb.WriteString(strconv.Itoa(l))
					sb.WriteRune('\n')
				}

				exactErr, _ := r.(error)

				err = &PanicError{
					err:   exactErr,
					stack: sb.String(),
				}
			}
		}()

		return f(ctx, p)
	}
}

func RequestConsumer[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan Request[P, R],
	processing RequestFunc[P, R],
) {
	if processing == nil {
		panic(ErrNilRequestProcessingFunc)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			for r := range inStream {
				r.Chan <- ValErrorPair[R]{
					Err: ctx.Err(),
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case req, more := <-inStream:
				if !more {
					return
				}

				select {
				case <-ctx.Done():
					req.Chan <- ValErrorPair[R]{
						Err: ctx.Err(),
					}

				default:
					req.Chan <- ToValError(
						interceptPanic82763487263(processing)(
							req.Ctx,
							req.P,
						),
					)
				}
			}
		}
	}()
}

func RequestConsumers[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	processing RequestFunc[P, R],
	inStreams ...<-chan Request[P, R],
) {
	for _, stream := range inStreams {
		RequestConsumer(ctx, wg, stream, processing)
	}
}

func RequestConsumersC[P any, R any](
	processing RequestFunc[P, R],
) ChainsFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan Request[P, R],
	) {
		RequestConsumers(ctx, wg, processing, inStreams...)
	}
}
