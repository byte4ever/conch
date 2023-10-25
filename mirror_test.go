package conch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMirrorHighThroughputC(t *testing.T) {
	goleak.VerifyNone(t, goleak.IgnoreCurrent())

	t.Run(
		"transmit values", func(t *testing.T) {

			subFunc := func(
				t *testing.T,
				replicas int,
				f func(
					replicas int,
					chains ChainsFunc[int],
				) ChainFunc[int],
			) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var wg sync.WaitGroup

				inStream := make(chan int)

				res := make([][]int, replicas)

				f(
					replicas,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStreams ...<-chan int,
					) {
						wg.Add(len(inStreams))

						for idx, stream := range inStreams {
							go func(
								i int,
								w *sync.WaitGroup,
								s <-chan int,
							) {
								defer w.Done()

								for {
									select {
									case v, more := <-s:
										if !more {
											return
										}

										res[i] = append(res[i], v)
									case <-ctx.Done():
										return
									}
								}
							}(idx, wg, stream)
						}

					},
				)(
					ctx,
					&wg,
					inStream,
				)

				require.Eventually(
					t,
					func() bool {
						inStream <- 10
						inStream <- 11
						inStream <- 12
						inStream <- 13

						return true
					},
					time.Second,
					time.Millisecond,
				)

				close(inStream)
				wgWait(t, &wg, time.Second, time.Millisecond)

				var expectedValues [][]int

				for i := 0; i < replicas; i++ {
					expectedValues = append(expectedValues, []int{10, 11, 12, 13})
				}

				require.Exactly(
					t,
					expectedValues,
					res,
				)
			}

			for _, replicas := range []int{1, 5, 7} {
				for _, f := range []func(
					replicas int,
					chains ChainsFunc[int],
				) ChainFunc[int]{
					MirrorsHighThroughputC[int],
					MirrorsLowLatencyC[int],
				} {
					subFunc(t, replicas, f)
				}
			}
		},
	)

	t.Run(
		"panics when replicas", func(t *testing.T) {
			for _, replicas := range []int{-1, 0} {
				for _, f := range []func(
					replicas int,
					chains ChainsFunc[int],
				) ChainFunc[int]{
					MirrorsHighThroughputC[int],
					MirrorsLowLatencyC[int],
				} {
					require.Panics(t, func() {
						f(replicas, nil)(nil, nil, nil)
					})
				}
			}
		},
	)
}
