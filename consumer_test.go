package conch

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
)

func TestConsumer(t *testing.T) {
	t.Parallel()

	t.Run(
		"closing input", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := NewMockLogger(t)

			d := NewMockDoer[int](t)
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

			l := NewMockLogger(t)
			l.On(
				"PanicError",
				"intercept panic",
				mock.MatchedBy(
					func(m map[string]any) bool {
						fmt.Println("================")
						pretty.Println(m)
						fmt.Println("================")

						return true
					},
				),
			)

			d := NewMockDoer[int](t)
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
					func(args mock.Arguments) {
						k := 0
						fmt.Println(120 / k)
						// panic("mocked panic")
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
}

func Test_A(t *testing.T) {
	localLogger, err := zap.NewDevelopment()
	require.NoError(t, err)
	ReplaceLogger(zapadapter.New(localLogger))
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inStream := make(chan int, 3)
	inStream <- 1
	inStream <- 2
	inStream <- 3
	close(inStream)

	Consumer(
		ctx,
		&wg,
		func(ctx context.Context, x int) {
			fmt.Println(123123123 / (x - 2))
		},
		inStream,
	)

	// cancel()
	require.Eventually(
		t, func() bool {
			wg.Wait()
			return true
		}, time.Second, time.Millisecond,
	)
}
