package dirty

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

// f1 is paniking when x = 2
//
//go:noinline
func f1(x int) {
	f2(x)
}

//go:noinline
func f2(x int) {
	f3(x)
}

//go:noinline
func f3(x int) {
	f4(x)
}

//go:noinline
func f4(x int) {
	x += 1
	x -= 1
	fmt.Println(123123123 / (x - 2))
}

func Test_consumer(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := dirty.NewMockLogger(t)

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			consumer(
				ctx,
				&wg,
				0,
				l,
				d.Execute,
				inStream,
			)

			// cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"panic processing", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := dirty.NewMockLogger(t)
			l.On(
				"Error",
				"conch intercepts panic",
				mock.MatchedBy(
					func(m map[string]any) bool {
						return assert.Contains(
							// error key MUST be present
							t,
							m,
							"error",
						) && assert.Equal(
							// error message MUST match original message
							t,
							"runtime error: integer divide by zero",
							m["error"],
						) && assert.Contains(
							// stack key MUST be present
							t,
							m,
							"stack",
						) && assert.Contains(
							// MUST contain the function name that panics
							t,
							m["stack"],
							".f4",
						) && assert.Contains(
							// MUST contain the file name and  panics locaion
							t,
							m["stack"],
							"/consumer_test.go:37",
						)
					},
				),
			)

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Run(
					func(mock.Arguments) {
						f1(2)
					},
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			consumer(
				ctx,
				&wg,
				0,
				l,
				d.Execute,
				inStream,
			)

			// cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"cancel kills the processor", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := dirty.NewMockLogger(t)
			d := dirty.NewMockDoer[int](t)

			inStream := make(chan int)
			defer close(inStream)

			consumer(
				ctx,
				&wg,
				0,
				l,
				d.Execute,
				inStream,
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

	t.Run(
		"panicking when doer is nil", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := dirty.NewMockLogger(t)

			inStream := make(chan int)
			defer close(inStream)

			require.PanicsWithValue(
				t,
				ErrNilConsumerDoer,
				func() {
					consumer(
						ctx,
						&wg,
						0,
						l,
						nil,
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

const consumerPoolSize = 10

func Test_Consumer(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			Consumer(ctx, &wg, 0, d.Execute, inStream)

			// cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"cancel kills the processor", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)

			inStream := make(chan int)
			defer close(inStream)

			Consumer(ctx, &wg, 0, d.Execute, inStream)

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

func Test_ConsumerPool(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			BalanceC(
				consumerPoolSize,
				ConsumersC(
					d.Execute,
				),
			)(ctx, &wg, inStream)

			// cancel()
			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"cancel kills the processor", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)

			inStream := make(chan int)
			defer close(inStream)

			BalanceC(
				consumerPoolSize,
				ConsumersC(
					d.Execute,
				),
			)(ctx, &wg, inStream)

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

func Test_ConsumerPoolC(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			BalanceC(
				consumerPoolSize,
				ConsumersC(
					d.Execute,
				),
			)(ctx, &wg, inStream)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, time.Second, time.Millisecond,
			)
		},
	)

	t.Run(
		"cancel kills the processor", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)

			inStream := make(chan int)
			defer close(inStream)

			BalanceC(
				consumerPoolSize,
				ConsumersC(
					d.Execute,
				),
			)(ctx, &wg, inStream)

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

func Test_ConsumerC(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)
			d.
				On(
					"Execute",
					ctx,
					1,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					2,
				).
				Return(nil).
				Once()
			d.
				On(
					"Execute",
					ctx,
					3,
				).
				Return(nil).
				Once()

			inStream := make(chan int, 3)
			inStream <- 1
			inStream <- 2
			inStream <- 3
			close(inStream)

			ConsumerC(0, d.Execute)(
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
		},
	)

	t.Run(
		"cancel kills the processor", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			d := dirty.NewMockDoer[int](t)

			inStream := make(chan int)
			defer close(inStream)

			ConsumerC(0, d.Execute)(
				ctx,
				&wg,
				inStream,
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