package dirty

/*
import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func doit(_ context.Context, p int) (int, error) {
	if p%2 == 0 {
		return 0, ErrMocked
	}

	return p*17 + 11, nil
}

func Test_ReqtestTogether(t *testing.T) {
	t.Parallel()

	t.Run(
		"concurrent single consumer", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			requester := RequesterC(
				ctx,
				&wg,
				RequestConsumerC(
					func(ctx context.Context, p int) (int, error) {
						if p%2 == 0 {
							return 0, ErrMocked
						}
						return p*17 + 11, nil
					},
				),
			)

			var wgR sync.WaitGroup
			var wgStart sync.WaitGroup

			const ConcurrentRequests = 1000

			wgR.Add(ConcurrentRequests)
			wgStart.Add(ConcurrentRequests)
			for i := 0; i < ConcurrentRequests; i++ {
				go func(i int) {
					defer wgR.Done()
					defer func() {
						fmt.Printf("wgR.Done() %d\n", i)
					}()

					wgStart.Done()
					r, err := requester(context.Background(), i)
					if i%2 == 0 {
						require.ErrorIs(t, err, ErrMocked)
						return
					}
					require.Equal(t, i*17+11, r)
				}(i)
			}

			require.Eventually(
				t, func() bool {
					wgStart.Wait()
					return true
				}, time.Second, time.Millisecond,
			)

			require.Eventually(
				t, func() bool {
					wgR.Wait()
					return true
				}, time.Second, time.Millisecond,
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
		"non concurrent single consumer", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			requester := RequesterC(
				ctx,
				&wg,
				RequestConsumerC(
					doit,
				),
			)

			const ConcurrentRequests = 1_000_000

			for i := 0; i < ConcurrentRequests; i++ {
				r, err := requester(context.Background(), i)
				if i%2 == 0 {
					require.ErrorIs(t, err, ErrMocked)
					continue
				}
				require.Equal(t, i*17+11, r)
			}

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
		"concurrent pool consumer", func(t *testing.T) {
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			requester := RequesterC(
				ctx,
				&wg,
				RequestConsumerPoolC(
					doit,
					32,
				),
			)

			time.Sleep(time.Second)

			var wgR sync.WaitGroup
			var wgStart sync.WaitGroup

			const ConcurrentRequests = 1_000_000

			wgR.Add(ConcurrentRequests)
			wgStart.Add(ConcurrentRequests)
			for i := 0; i < ConcurrentRequests; i++ {
				go func(i int) {
					defer wgR.Done()
					wgStart.Done()
					r, err := requester(context.Background(), i)
					if i%2 == 0 {
						require.ErrorIs(t, err, ErrMocked)
						return
					}
					require.Equal(t, i*17+11, r)
				}(i)
			}

			require.Eventually(
				t, func() bool {
					wgStart.Wait()
					return true
				}, 10*time.Second, time.Millisecond,
			)

			require.Eventually(
				t, func() bool {
					wgR.Wait()
					return true
				}, 10*time.Second, time.Millisecond,
			)

			cancel()

			require.Eventually(
				t, func() bool {
					wg.Wait()
					return true
				}, 10*time.Second, time.Millisecond,
			)
		},
	)
}
*/
