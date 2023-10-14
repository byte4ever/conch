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
	ErrInvalidConsumerPoolCount = errors.New("invalid consumer pool count")
	ErrNilConsumerDoer          = errors.New("nil consumer doer")
	ErrNilConsumerPoolDoer      = errors.New("nil consumer pool doer doer")
)

func intercept[T any](
	logger Logger,
	f Doer[T],
) Doer[T] {
	const (
		MaxCallerStakSize = 128
		StringBufferSize  = 2048
	)

	return func(ctx context.Context, t T) {
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

				err, _ := r.(error)
				logger.Error(
					"conch intercepts panic",
					map[string]any{
						"error": err.Error(),
						"stack": sb.String(),
					},
				)
			}
		}()

		f(ctx, t)
	}
}

// Consumer spawn a single consumer for the input stream.
//
// It's shut down when the input stream is closed or the context is
// canceled and wg is used to wait for the consumer to finish when shutdown
// condition is met.
//
// Panics that may occur during doer execution will be caught and
// logged on global logger.
func Consumer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	f Doer[T],
	inStream <-chan T,
) {
	consumer(ctx, wg, logger, f, inStream)
}

func consumer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	l Logger,
	f Doer[T],
	inStream <-chan T,
) {
	if f == nil {
		panic(ErrNilConsumerDoer)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		fi := intercept(l, f)

		for {
			select {
			case <-ctx.Done():
				return

			case e, more := <-inStream:
				if !more {
					return
				}

				fi(ctx, e)
			}
		}
	}()
}

// ConsumerPool spawns n concurrent  consumers for the input stream.
//
// All of them are shut down when the input stream is closed or the context is
// canceled and wg is used to wait for all consumers to finish when shutdown
// condition is met.
//
// Panics that may occur during doer execution will be caught and
// logged on global logger.
func ConsumerPool[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	f Doer[T],
	inStream <-chan T,
) {
	consumerPool(ctx, wg, logger, count, f, inStream)
}

func consumerPool[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	l Logger,
	count int,
	f Doer[T],
	inStream <-chan T,
) {
	if count <= 0 {
		panic(ErrInvalidConsumerPoolCount)
	}

	if f == nil {
		panic(ErrNilConsumerPoolDoer)
	}

	for i := 0; i < count; i++ {
		consumer(ctx, wg, l, f, inStream)
	}
}

func ConsumerC[T any](
	f Doer[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		Consumer(ctx, wg, f, inStream)
	}
}

func ConsumerPoolC[T any](
	count int,
	f Doer[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		ConsumerPool(ctx, wg, count, f, inStream)
	}
}
