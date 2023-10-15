package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func wgWait(
	t *testing.T,
	wg *sync.WaitGroup,
	waitFor time.Duration,
	tick time.Duration,
) {
	t.Helper()
	require.Eventually(
		t, func() bool {
			wg.Wait()
			return true
		},
		waitFor,
		tick,
	)
}

func scaffoldTest[I any](
	t *testing.T,
	f func(t *testing.T, inStream chan I, wg *sync.WaitGroup),
) {
	t.Helper()

	var wg sync.WaitGroup

	inStream := make(chan I)
	f(t, inStream, &wg)

	wgWait(t, &wg, time.Second, time.Millisecond)
}

func isClosed[T any](t *testing.T, c <-chan T) {
	t.Helper()

	_, more := <-c
	require.False(t, more)
}

func isClosing[T any](
	t *testing.T,
	c <-chan T,
	waitFor time.Duration,
	tick time.Duration,
) {
	t.Helper()
	require.Eventually(
		t, func() bool {
			var cnt int

			for range c {
				cnt++
			}

			t.Log("read", cnt, "items before closing")
			return true
		},
		waitFor,
		tick,
	)
}

func generator[T any](
	ctx context.Context,
	f func(
		ctx context.Context,
		n uint64,
	) (T, bool),
) chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		var cnt uint64

		for {
			select {
			case <-ctx.Done():
				return
			default:
				e, more := f(ctx, cnt)

				if !more {
					return
				}

				cnt++

				select {
				case <-ctx.Done():
					return
				case outStream <- e:
				}
			}
		}
	}()

	return outStream
}

func GeneratorProducer[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	f func(
		ctx context.Context,
		n uint64,
	) (T, bool),
	chain ChainFunc[T],
) {
	chain(ctx, wg, generator(ctx, f))
}

func BlockingSink[T any]() ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		<-ctx.Done()
	}
}
