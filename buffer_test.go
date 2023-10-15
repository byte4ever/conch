package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestBuffer(t *testing.T) {
	t.Parallel()

	t.Run(
		"ensure buffering is working",
		func(t *testing.T) {
			// having a forever blocking consumer un the chain and count number
			// of items actually buffered.

			const nbItems = 10

			var wg sync.WaitGroup

			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			GeneratorProducer(
				ctx,
				&wg,
				func(
					ctx context.Context,
					n uint64,
				) (uint64, bool) {
					if n != nbItems {
						return n, true
					}
					require.Equal(t, uint64(nbItems), n)
					cancel()
					return 0, false
				},
				BufferC(
					nbItems,
					BlockingSink[uint64](),
				),
			)
		},
	)

	t.Run(
		"ensure buffering is properly working",
		func(t *testing.T) {
			// having a counting consumer ending the chain and count number
			// of items actually buffered.

			const nbItems = 100

			var wg sync.WaitGroup
			var maxReached uint64
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			require.Eventually(
				t,
				func() bool {
					GeneratorProducer(
						ctx,
						&wg,
						func(
							ctx context.Context,
							n uint64,
						) (uint64, bool) {
							if n != nbItems {
								//t.Log("producer", n)
								return n, true
							}
							return 0, false
						},
						BufferC(
							5,
							ConsumerC(0, func(ctx context.Context, id int, cnt uint64) {
								//t.Log("consumer", id)
								maxReached = cnt
							}),
						),
					)

					wg.Wait()

					return assert.Equal(
						t,
						uint64(nbItems)-1,
						maxReached,
					)
				},
				time.Second,
				20*time.Millisecond,
			)
		},
	)

	t.Run(
		"ensure context canceling is properly working - 1",
		func(t *testing.T) {
			// by feeding half the buffer then canceling

			const nbItems = 100

			var wg sync.WaitGroup
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			require.Eventually(
				t,
				func() bool {
					GeneratorProducer(
						ctx,
						&wg,
						func(
							ctx context.Context,
							n uint64,
						) (uint64, bool) {
							if n != nbItems {
								//t.Log("producer", n)
								if n == nbItems/2 {
									cancel()
								}
								return n, true
							}
							return 0, false
						},
						BufferC(
							nbItems/3,
							ConsumerC(0, func(
								ctx context.Context,
								id int,
								cnt uint64,
							) {
								//t.Log("consumer", id)
							}),
						),
					)

					wg.Wait()

					return true
				},
				time.Second,
				20*time.Millisecond,
			)
		},
	)

	t.Run(
		"ensure context canceling is properly working - 2",
		func(t *testing.T) {
			// by feeding half the buffer then canceling

			const nbItems = 100

			var wg sync.WaitGroup
			t.Parallel()

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			require.Eventually(
				t,
				func() bool {
					GeneratorProducer(
						ctx,
						&wg,
						func(
							ctx context.Context,
							n uint64,
						) (uint64, bool) {
							if n != nbItems {
								//t.Log("producer", n)
								return n, true
							}
							return 0, false
						},
						BufferC(
							nbItems/2,
							ConsumerC(0, func(
								ctx context.Context,
								id int,
								cnt uint64,
							) {
								//t.Log("consumer", cnt)
								if cnt == nbItems/4 {
									cancel()
								}
							}),
						),
					)

					wg.Wait()

					return true
				},
				time.Second,
				20*time.Millisecond,
			)
		},
	)
}
