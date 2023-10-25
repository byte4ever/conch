package conch

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestInjectorC(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	t.Run(
		"transmit values", func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			var collectedValues []int

			var allDone sync.WaitGroup
			allDone.Add(100)

			f := InjectorC(
				ctx,
				&wg,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStreams ...<-chan int,
				) {
					wg.Add(len(inStreams))

					for _, inStream := range inStreams {
						go func(s <-chan int) {
							defer wg.Done()

							for {
								select {
								case v, more := <-s:
									if !more {
										return
									}
									collectedValues = append(collectedValues, v)
									allDone.Done()
								case <-ctx.Done():
									return
								}
							}
						}(inStream)
					}
				},
			)

			expectedValues := make([]int, 100)
			for i := 0; i < 100; i++ {
				require.Eventually(t, func() bool {
					return assert.NoError(t, f(ctx, i))
				}, time.Second, time.Millisecond)

				expectedValues[i] = i
			}

			require.Eventually(t, func() bool {
				allDone.Wait()
				return true
			}, time.Second, time.Millisecond)

			require.Exactly(t, expectedValues, collectedValues)

			cancel()
			wg.Wait()
		},
	)

	t.Run(
		"cancel innerctx", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			f := InjectorC(
				ctx,
				&wg,
				MultiplexC(10,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStreams ...<-chan int,
					) {
						wg.Add(len(inStreams))

						for _, inStream := range inStreams {
							go func(s <-chan int) {
								defer wg.Done()
								for {
									select {
									case _, more := <-s:
										if !more {
											return
										}
									case <-ctx.Done():
										return
									}
								}
							}(inStream)
						}
					},
				),
			)

			innerCtx, innerCancel := context.WithCancel(context.Background())
			defer innerCancel()

			innerCancel()

			require.ErrorIs(t, f(innerCtx, 1234), context.Canceled)

			cancel()
			wg.Wait()
		},
	)

	t.Run(
		"cancel main ctx", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			f := InjectorC(
				ctx,
				&wg,
				MultiplexC(10,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStreams ...<-chan int,
					) {
						wg.Add(len(inStreams))

						for _, inStream := range inStreams {
							go func(s <-chan int) {
								defer wg.Done()
								for {
									select {
									case _, more := <-s:
										if !more {
											return
										}
									case <-ctx.Done():
										return
									}
								}
							}(inStream)
						}
					},
				),
			)

			innerCtx, innerCancel := context.WithCancel(context.Background())
			defer innerCancel()

			cancel()
			require.ErrorIs(t, f(innerCtx, 1234), context.Canceled)

			wg.Wait()
		},
	)

	t.Run(
		"stress coverage", func(t *testing.T) {
			const TestCount = 10_000

			for i := 0; i < TestCount; i++ {
				bolos(t)
			}
		},
	)
}

func bolos(t *testing.T) {
	t.Helper()

	const testDuration = 10 * time.Millisecond

	ctx, cancel := context.WithTimeout(
		context.Background(),
		testDuration,
	)

	defer cancel()

	var wg sync.WaitGroup

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	f := InjectorC(
		ctx,
		&wg,
		func(
			ctx context.Context,
			wg *sync.WaitGroup,
			inStreams ...<-chan int,
		) {
			wg.Add(len(inStreams))

			for _, inStream := range inStreams {
				go func(s <-chan int) {
					defer wg.Done()

					for {
						select {
						case <-ctx.Done():
							return
						}
					}
				}(inStream)
			}
		},
	)

	for {
		randomTimeout := time.Duration(rnd.Intn(int(testDuration) / 4))
		innerCtx, innerCancel := context.WithTimeout(
			context.Background(),
			randomTimeout,
		)

		err := f(innerCtx, 100)
		if err != nil {
			innerCancel()
			break
		}

		innerCancel()
	}

	cancel()
	wg.Wait()
}
