package conch

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func consumerProvider[P, R any](
	t *testing.T,
	functor func(P) (R, error),
	concurrency int,
) func(
	ctx context.Context,
	wg *sync.WaitGroup,
	inStreams ...<-chan Request[P, R],
) {
	t.Helper()

	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStreams ...<-chan Request[P, R],
	) {
		require.Len(t, inStreams, concurrency)
		wg.Add(concurrency)

		for idx, inStream := range inStreams {
			go func(inStream <-chan Request[P, R], idx int) {
				defer wg.Done()

				for request := range inStream {
					select {
					case <-ctx.Done():
						request.Chan <- ValErrorPair[R]{
							Err: ctx.Err(),
						}

						return

					case <-request.Ctx.Done():
						request.Chan <- ValErrorPair[R]{
							Err: request.Ctx.Err(),
						}

					case request.Chan <- ToValError(functor(request.P)):
					}
				}
			}(inStream, idx)
		}
	}
}

func TestRequestersC(t *testing.T) { //nolint:maintidx // dgas
	t.Run(
		"run requests", func(t *testing.T) {
			const TestCount = 100

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			const concurrency = 10
			requesters := RequestersC(
				ctx,
				&wg,
				concurrency,
				consumerProvider(t, lcg, concurrency),
			)

			var runGroup sync.WaitGroup
			runGroup.Add(TestCount)

			for i := uint64(0); i < TestCount; i++ {
				requestIdx := int(i) % concurrency
				go func(wg *sync.WaitGroup, v uint64, requestIdx int) {
					defer wg.Done()

					val, err := requesters[requestIdx](
						context.Background(),
						v,
					)
					expectedVal, expectedError := lcg(v)

					require.Equal(t, val, expectedVal)
					require.ErrorIs(t, err, expectedError)
				}(
					&runGroup,
					rand.Uint64(),
					requestIdx,
				)
			}

			require.Eventually(t, func() bool {
				runGroup.Wait()
				return true
			}, time.Second, 10*time.Millisecond)

			cancel()

			require.Eventually(t, func() bool {
				wg.Wait()
				return true
			}, time.Second, 10*time.Millisecond)

		},
	)

	t.Run(
		"stress global context", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			for i := 0; i < 1000; i++ {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)

				var wg sync.WaitGroup

				const concurrency = 10
				requesters := RequestersC(
					ctx,
					&wg,
					concurrency,
					consumerProvider(t, lcg, concurrency),
				)

			again:
				select {
				case <-ctx.Done():
					goto out
				default:
					param := rand.Uint64()

					expectedResult, expectedErr := lcg(param)

					result, err := requesters[rand.Intn(concurrency)](context.Background(), param)

					if errors.Is(err, context.DeadlineExceeded) {
						require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
						require.Equal(t, result, uint64(0))
						goto out
					}

					require.ErrorIs(t, err, expectedErr)
					require.Equal(t, result, expectedResult)

					goto again
				}

			out:
				require.Eventually(t, func() bool {
					wg.Wait()
					return true
				}, time.Second, 10*time.Millisecond)
				cancel()
			}
		},
	)

	t.Run(
		"stress request context", func(t *testing.T) {
			const TestCount = 10_000

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			const concurrency = 10
			requesters := RequestersC(
				ctx,
				&wg,
				concurrency,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStreams ...<-chan Request[uint64, uint64],
				) {
					require.Len(t, inStreams, concurrency)

					wg.Add(concurrency)

					for _, inStream := range inStreams {
						go func(s <-chan Request[uint64, uint64]) {
							defer wg.Done()

							for request := range s {
								select {
								case <-ctx.Done():
									request.Chan <- ValErrorPair[uint64]{
										Err: ctx.Err(),
									}

									return

								case <-request.Ctx.Done():
									request.Chan <- ValErrorPair[uint64]{
										Err: request.Ctx.Err(),
									}
								}
							}
						}(inStream)
					}

				},
			)

			var runGroup sync.WaitGroup
			runGroup.Add(TestCount)

			for i := uint64(0); i < TestCount; i++ {
				requestIdx := int(i) % concurrency
				go func(wg *sync.WaitGroup, v uint64, requestIdx int) {
					defer wg.Done()

					locCtx, cancel := context.WithTimeout(
						context.Background(),
						10*time.Millisecond,
					)
					defer cancel()

					val, err := requesters[requestIdx](
						locCtx,
						v,
					)

					require.ErrorIs(t, err, context.DeadlineExceeded)
					require.Equal(t, val, uint64(0))
				}(
					&runGroup,
					rand.Uint64(),
					requestIdx,
				)
			}

			require.Eventually(t, func() bool {
				runGroup.Wait()
				return true
			}, time.Second, 10*time.Millisecond)

			time.Sleep(20 * time.Millisecond)

			cancel()
			require.Eventually(t, func() bool {
				wg.Wait()
				return true
			}, time.Second, 10*time.Millisecond)
		},
	)

	t.Run(
		"single run requests", func(t *testing.T) {
			const TestCount = 100

			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			requester := RequesterC(
				ctx,
				&wg,
				consumerProvider(t, lcg, 1),
			)

			var runGroup sync.WaitGroup
			runGroup.Add(TestCount)

			for i := uint64(0); i < TestCount; i++ {
				go func(wg *sync.WaitGroup, v uint64) {
					defer wg.Done()

					val, err := requester(
						context.Background(),
						v,
					)
					expectedVal, expectedError := lcg(v)

					require.Equal(t, val, expectedVal)
					require.ErrorIs(t, err, expectedError)
				}(
					&runGroup,
					rand.Uint64(),
				)
			}

			require.Eventually(t, func() bool {
				runGroup.Wait()
				return true
			}, time.Second, 10*time.Millisecond)

			cancel()

			require.Eventually(t, func() bool {
				wg.Wait()
				return true
			}, time.Second, 10*time.Millisecond)

		},
	)
}
