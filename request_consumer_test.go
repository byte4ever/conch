package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_interceptPanic(t *testing.T) {
	t.Run(
		"when ok", func(t *testing.T) {
			v, err := interceptPanic(func(ctx context.Context, p int) (int, error) {
				return p*13 + 1, nil
			})(context.Background(), 100)

			require.NoError(t, err)
			require.Equal(t, 1301, v)
		},
	)

	t.Run(
		"when error", func(t *testing.T) {
			v, err := interceptPanic(func(ctx context.Context, p int) (int, error) {
				return 0, ErrMocked
			})(context.Background(), 100)

			require.ErrorIs(t, err, ErrMocked)
			require.Zero(t, v)
		},
	)

	t.Run(
		"when panicking", func(t *testing.T) {
			v, err := interceptPanic(
				func(
					ctx context.Context,
					p int,
				) (
					int,
					error,
				) {
					panic("test")
					return 0, nil
				},
			)(
				context.Background(),
				100,
			)

			require.Zero(t, v)

			var errPanic *PanicError
			require.ErrorAs(t, err, &errPanic)
		},
	)

}

func lcgRequest(
	_ context.Context,
	p uint64,
) (
	uint64,
	error,
) {
	return lcg(p)
}

func TestRequestConsumersC(t *testing.T) {
	t.Run(
		"answer requests", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var group sync.WaitGroup

			inStream := make(chan Request[uint64, uint64])

			RequestConsumersC(lcgRequest)(ctx, &group, inStream)

			respChan := make(chan ValErrorPair[uint64])
			req := Request[uint64, uint64]{
				P:    100,
				Ctx:  context.Background(),
				Chan: respChan,
			}

			require.Eventually(t, func() bool {
				inStream <- req
				return true
			}, time.Second, 10*time.Millisecond)

			require.Eventually(t, func() bool {
				valError := <-respChan
				expectedVal, expectedErr := lcg(100)
				require.Equal(
					t,
					ValErrorPair[uint64]{
						V:   expectedVal,
						Err: expectedErr,
					},
					valError,
				)
				return true
			}, time.Second, 10*time.Millisecond)

			close(inStream)

			cancel()
			wgWait(t, &group, time.Second, 10*time.Millisecond)
		},
	)
}
