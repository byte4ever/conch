package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestCounting(t *testing.T) { //nolint:maintidx //dgas
	t.Parallel()

	const testCount = 20

	t.Run(
		"input is closing", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx := context.Background()

			s := Counting(
				ctx,
				fakeStream2(
					func(i int) struct{} {
						return struct{}{}
					},
					testCount,
				),
			)

			for i := 0; i < testCount; i++ {
				idx := i

				require.Eventually(
					t, func() bool {
						k := assert.Equal(
							t, idx, <-s,
						)

						b := k
						return b
					},
					time.Second,
					time.Millisecond,
				)
			}

			isClosed(t, s)
		},
	)

	t.Run(
		"context is canceled", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s := Counting(
				ctx,
				fakeStream2(
					func(i int) struct{} {
						return struct{}{}
					},
					testCount,
				),
			)

			for i := 0; i < 10; i++ {
				idx := i

				require.Eventually(
					t, func() bool {
						k := assert.Equal(
							t, idx, <-s,
						)

						b := k
						return b
					},
					time.Second,
					time.Millisecond,
				)
			}

			cancel()
			isClosed(t, s)
		},
	)
}

func TestCountingC(t *testing.T) { //nolint:maintidx //dgas
	t.Parallel()

	const testCount = 20

	t.Run(
		"closed input", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx := context.Background()

			var wg sync.WaitGroup

			mockedDoer := NewMockDoer[int](t)
			for i := 0; i < testCount; i++ {
				mockedDoer.
					On("Execute", ctx, i).
					Return(nil).
					Once()
			}

			CountingC[struct{}](
				ConsumerC(0, mockedDoer.Execute),
			)(
				ctx,
				&wg,
				fakeStream2(
					func(i int) struct{} {
						return struct{}{}
					},
					testCount,
				),
			)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				},
				time.Second,
				time.Millisecond,
			)
		},
	)

	t.Run(
		"canceled context", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockedDoer := NewMockDoer[int](t)
			for i := 0; i < 10; i++ {
				mockedDoer.
					On("Execute", ctx, i).
					Return(nil).
					Once()
			}

			mockedDoer.On("Execute", ctx, mock.AnythingOfType("int")).
				Run(
					func(args mock.Arguments) {
						cancel()
					},
				).
				Return(nil).
				Once()

			CountingC[struct{}](
				ConsumerC(0, mockedDoer.Execute),
			)(
				ctx,
				&wg,
				fakeStream2(
					func(i int) struct{} {
						return struct{}{}
					},
					testCount,
				),
			)

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				},
				time.Second,
				time.Millisecond,
			)
		},
	)
}
