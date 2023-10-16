package dirty

/*
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	ErrMocked = errors.New("mocked error")
)

func Test_RequestConsumerC(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			res := make(map[int]ValErrorPair[int])

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := NewMockRequestFunc[int, int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(10, error(nil)).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(0, ErrMocked).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(30, error(nil)).
				Once()
			d.
				On(
					"Execute",
					ctx,
					1000,
				).
				Run(
					func(args mock.Arguments) {
						x := 0
						x += 1
						x -= 1
						fmt.Println("x", 1/x)
					},
				).
				Return().
				Once()

			inStream := make(chan Request[int, int], 4)
			inStream <- Request[int, int]{
				P: 1000,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[1000] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[1000] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 1,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[1] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[1] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 2,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[2] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[2] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 3,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[3] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[3] = v
				// },
			}
			close(inStream)

			RequestConsumerC[int, int](
				d.Execute,
			)(
				ctx,
				&wg,
				inStream,
			)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			require.Len(t, res, 4)

			require.Contains(t, res, 1)
			require.Equal(
				t,
				ValErrorPair[int]{
					V: 10,
				},
				res[1],
			)

			require.Contains(t, res, 2)
			require.Equal(
				t,
				ValErrorPair[int]{
					Err: ErrMocked,
				},
				res[2],
			)

			require.Contains(t, res, 3)
			require.Equal(
				t,
				ValErrorPair[int]{
					V: 30,
				},
				res[3],
			)

			require.Contains(t, res, 1000)
			require.IsType(t, &PanicError{}, res[1000].Err)
			var perr *PanicError

			require.ErrorAs(t, res[1000].Err, &perr)
			require.ErrorContains(
				t,
				perr.err,
				"runtime error: integer divide by zero",
			)
			require.Contains(
				t,
				perr.stack,
				"request_consumer_test.go:71",
			)
		},
	)

	t.Run(
		"cancel ctx", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := NewMockRequestFunc[int, int](t)
			inStream := make(chan Request[int, int])

			RequestConsumerC[int, int](
				d.Execute,
			)(
				ctx,
				&wg,
				inStream,
			)

			cancel()

			close(inStream) // this is required to enable draining

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"panicking when processing func is nil", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			inStream := make(chan Request[int, int])
			defer close(inStream)

			require.PanicsWithValue(
				t, ErrNilRequestProcessingFunc, func() {
					RequestConsumerC[int, int](
						nil,
					)(
						ctx,
						&wg,
						inStream,
					)
				},
			)

			cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)
}

func Test_RequestConsumerPoolC(t *testing.T) {
	const RequesConsummerPooSize = 2

	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			res := make(map[int]ValErrorPair[int])

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := NewMockRequestFunc[int, int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(10, error(nil)).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(0, ErrMocked).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(30, error(nil)).
				Once()
			d.
				On(
					"Execute",
					ctx,
					1000,
				).
				Run(
					func(args mock.Arguments) {
						x := 0
						x += 1
						x -= 1
						fmt.Println("x", 1/x)
					},
				).
				Return().
				Once()

			inStream := make(chan Request[int, int], 4)
			inStream <- Request[int, int]{
				P: 1000,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[1000] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[1000] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 1,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[1] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[1] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 2,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[2] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[2] = v
				// },
			}
			inStream <- Request[int, int]{
				P: 3,
				NakedReturn: func(provider ValErrorPairProvider[int]) {
					res[3] = provider(ctx)
				},
				// Return: func(
				// 	ctx context.Context,
				// 	v ValErrorPair[int],
				// ) {
				// 	res[3] = v
				// },
			}

			close(inStream)

			RequestConsumerPoolC[int, int](
				d.Execute,
				RequesConsummerPooSize,
			)(
				ctx,
				&wg,
				inStream,
			)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			require.Len(t, res, 4)

			require.Contains(t, res, 1)
			require.Equal(
				t,
				ValErrorPair[int]{
					V: 10,
				},
				res[1],
			)

			require.Contains(t, res, 2)
			require.Equal(
				t,
				ValErrorPair[int]{
					Err: ErrMocked,
				},
				res[2],
			)

			require.Contains(t, res, 3)
			require.Equal(
				t,
				ValErrorPair[int]{
					V: 30,
				},
				res[3],
			)

			require.Contains(t, res, 1000)
			require.IsType(t, &PanicError{}, res[1000].Err)
			var perr *PanicError

			require.ErrorAs(t, res[1000].Err, &perr)
			require.ErrorContains(
				t,
				perr.err,
				"runtime error: integer divide by zero",
			)
			require.Contains(
				t,
				perr.stack,
				"request_consumer_test.go:313",
			)
		},
	)

	t.Run(
		"cancel ctx", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := NewMockRequestFunc[int, int](t)
			inStream := make(chan Request[int, int])

			RequestConsumerPoolC[int, int](
				d.Execute,
				RequesConsummerPooSize,
			)(
				ctx,
				&wg,
				inStream,
			)

			cancel()

			close(inStream) // this is required to enable draining

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"panicking when processing func is nil", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			inStream := make(chan Request[int, int])
			defer close(inStream)

			require.PanicsWithValue(
				t, ErrNilRequestProcessingFunc, func() {
					RequestConsumerPoolC[int, int](
						nil,
						RequesConsummerPooSize,
					)(
						ctx,
						&wg,
						inStream,
					)
				},
			)

			cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)
}
*/
