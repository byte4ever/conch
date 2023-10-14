package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

type NumberErr int

func (n NumberErr) Error() string {
	return ""
}

func TestSeparateC(t *testing.T) {
	var (
		values  []int
		_errors []error
	)

	defer goleak.VerifyNone(
		t,
		goleak.IgnoreCurrent(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	inStream := make(chan ValErrorPair[int])

	SeparateC[int](
		func(
			ctx context.Context,
			group *sync.WaitGroup,
			inStream <-chan int,
		) {
			group.Add(1)

			go func() {
				group.Done()

				for {
					select {
					case <-ctx.Done():
						return
					case v, more := <-inStream:
						if !more {
							return
						}
						values = append(values, v)
					}
				}
			}()
		},
		func(
			ctx context.Context,
			group *sync.WaitGroup,
			inStream <-chan error,
		) {
			group.Add(1)

			go func() {
				group.Done()

				for {
					select {
					case <-ctx.Done():
						return
					case v, more := <-inStream:
						if !more {
							return
						}
						_errors = append(_errors, v)
					}
				}
			}()
		},
	)(ctx, &wg, inStream)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			V: 1,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			V: 2,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			Err: NumberErr(1),
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			Err: NumberErr(2),
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			V: 3,
		}

		return true
	}, time.Second, time.Millisecond)

	require.Eventually(t, func() bool {
		inStream <- ValErrorPair[int]{
			Err: NumberErr(3),
		}

		return true
	}, time.Second, time.Millisecond)

	close(inStream)
	wg.Wait()

	require.Equal(
		t,
		[]int{
			1,
			2,
			3,
		},
		values,
	)
	require.Equal(
		t,
		[]error{
			NumberErr(1),
			NumberErr(2),
			NumberErr(3),
		},
		_errors,
	)
}
