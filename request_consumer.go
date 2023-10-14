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

func interceptRequest[P, R any](
	f RequestFunc[P, R],
) RequestFunc[P, R] {
	const (
		MaxCallerStakSize = 128
		StringBufferSize  = 2048
	)

	return func(ctx context.Context, p P) (res R, err error) {
		defer func() {
			if r := recover(); r != nil {
				var sb strings.Builder

				sb.Grow(StringBufferSize)

				stk := make([]uintptr, MaxCallerStakSize)
				n := runtime.Callers(4, stk)
				stk = stk[:n]

				for _, pc := range stk {
					f := runtime.FuncForPC(pc)
					file, l := f.FileLine(pc)
					name := f.Name()

					if strings.Contains(name, "conch.intercept[...].func1") {
						break
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

		fi := interceptRequest(processing)

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
				case <-req.Ctx.Done():
					req.Chan <- ValErrorPair[R]{
						Err: req.Ctx.Err(),
					}
				default:
					r, err := fi(req.Ctx, req.P)
					req.Chan <- ValErrorPair[R]{
						V:   r,
						Err: err,
					}
				}
			}
		}
	}()
}

func RequestConsumerC[P any, R any](
	processing RequestFunc[P, R],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		RequestConsumer(
			ctx, wg, inStream, processing,
		)
	}
}

func RequestConsumerPool[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan Request[P, R],
	processing RequestFunc[P, R],
	count int,
) {
	for i := 0; i < count; i++ {
		RequestConsumer(
			ctx,
			wg,
			inStream,
			processing,
		)
	}
}

func RequestConsumerPoolC[P any, R any](
	processing RequestFunc[P, R],
	count int,
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		RequestConsumerPool(ctx, wg, inStream, processing, count)
	}
}
