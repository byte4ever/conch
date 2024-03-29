package conch

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMultiplexC(t *testing.T) {
	t.Run(
		"one to many", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			const count = 200
			const outCount = 10

			inStream := make(chan int, count)

			collected := make([][]int, outCount)

			MultiplexC(
				outCount,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream ...<-chan int,
				) {
					wg.Add(outCount)

					var ww sync.WaitGroup
					ww.Add(1)
					for i := 0; i < outCount; i++ {
						go func(c <-chan int, idx int) {
							defer wg.Done()

							ww.Wait()
							for val := range c {
								collected[idx] = append(collected[idx], val)
								runtime.Gosched()
							}
						}(inStream[i], i)
					}
					ww.Done()
				},
			)(ctx, &wg, inStream)

			for i := 0; i < count; i++ {
				inStream <- i
			}

			close(inStream)

			wgWait(t, &wg, time.Second, time.Millisecond)

			var allOfThem []int
			for _, ints := range collected {
				require.IsIncreasing(t, ints)
				allOfThem = append(allOfThem, ints...)
			}

			sort.Ints(allOfThem)
			for i, i2 := range allOfThem {
				require.Equal(t, i, i2)
			}
		},
	)

	t.Run(
		"many to many", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			const count = 200
			const outCount = 10
			const inCount = 5

			inStreams := make([]chan int, inCount)

			for i := 0; i < inCount; i++ {
				inStreams[i] = make(chan int, count)
			}

			collected := make([][]int, outCount)

			inStreams2 := make([]<-chan int, inCount)
			for i := 0; i < inCount; i++ {
				inStreams2[i] = inStreams[i]
			}

			MultiplexC(
				outCount,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream ...<-chan int,
				) {
					wg.Add(outCount)

					var ww sync.WaitGroup
					ww.Add(1)
					for i := 0; i < outCount; i++ {
						go func(c <-chan int, idx int) {
							defer wg.Done()

							ww.Wait()
							for val := range c {
								collected[idx] = append(collected[idx], val)
								runtime.Gosched()
							}
						}(inStream[i], i)
					}
					ww.Done()
				},
			)(ctx, &wg, inStreams2...)

			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < count; i++ {
				inStreams[rnd.Intn(inCount)] <- i
			}

			for i := 0; i < inCount; i++ {
				close(inStreams[i])
			}

			wgWait(t, &wg, time.Second, time.Millisecond)

			var allOfThem []int
			for _, ints := range collected {
				allOfThem = append(allOfThem, ints...)
			}

			sort.Ints(allOfThem)
			for i, i2 := range allOfThem {
				require.Equal(t, i, i2)
			}
		},
	)

	t.Run(
		"many to one", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var wg sync.WaitGroup

			const count = 200
			const inCount = 5

			inStreams := make([]chan int, inCount)

			for i := 0; i < inCount; i++ {
				inStreams[i] = make(chan int, count)
			}

			var collected []int

			inStreams2 := make([]<-chan int, inCount)
			for i := 0; i < inCount; i++ {
				inStreams2[i] = inStreams[i]
			}

			MultiplexC(
				1,
				func(
					ctx context.Context,
					wg *sync.WaitGroup,
					inStream ...<-chan int,
				) {
					wg.Add(1)

					var ww sync.WaitGroup

					ww.Add(1)

					go func(c <-chan int) {
						defer wg.Done()

						ww.Wait()
						for val := range c {
							collected = append(collected, val)
							runtime.Gosched()
						}
					}(inStream[0])

					ww.Done()
				},
			)(ctx, &wg, inStreams2...)

			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < count; i++ {
				inStreams[rnd.Intn(inCount)] <- i
			}

			for i := 0; i < inCount; i++ {
				close(inStreams[i])
			}

			wgWait(t, &wg, time.Second, time.Millisecond)

			sort.Ints(collected)
			for i, i2 := range collected {
				require.Equal(t, i, i2)
			}
		},
	)

	t.Run(
		"cancel context stress test", func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			const testSize = 100

			for x := 0; x < testSize; x++ {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					10*time.Millisecond,
				)

				var wg sync.WaitGroup

				const count = 400

				outCount := 1 + rand.Intn(7)
				inCount := 1 + rand.Intn(7)

				inStreams := make([]chan int, inCount)

				for i := 0; i < inCount; i++ {
					inStreams[i] = make(chan int, count)
				}

				inStreams2 := make([]<-chan int, inCount)
				for i := 0; i < inCount; i++ {
					inStreams2[i] = inStreams[i]
				}

				MultiplexC(
					outCount,
					func(
						ctx context.Context,
						wg *sync.WaitGroup,
						inStream ...<-chan int,
					) {
						wg.Add(outCount)

						var ww sync.WaitGroup
						ww.Add(1)

						for i := 0; i < outCount; i++ {
							go func(c <-chan int, idx int) {
								defer wg.Done()

								ww.Wait()

								for range c {
									runtime.Gosched()
								}
							}(inStream[i], i)
						}

						ww.Done()
					},
				)(ctx, &wg, inStreams2...)

				wg.Add(inCount)

				var wStart sync.WaitGroup

				wStart.Add(1)
				for i := 0; i < inCount; i++ {
					go func(s chan int, idx int) {
						defer wg.Done()
						defer close(s)

						wStart.Wait()
						for {
							select {
							case <-ctx.Done():
								return
							case s <- 0:
							}
						}

					}(inStreams[i], i)
				}

				wStart.Done()
				wgWait(t, &wg, 10*time.Second, time.Millisecond)
				cancel()
			}
		},
	)

	t.Run(
		"panics when no input streams", func(t *testing.T) {
			require.PanicsWithValue(
				t,
				multiplexNoInputStreamMsg,
				func() {
					MultiplexC[any](10, nil)(nil, nil)
				},
			)
		},
	)

	t.Run(
		"panics when count is not strictly positive", func(t *testing.T) {
			require.PanicsWithValue(
				t,
				fmt.Sprintf(multiplexInvalidCountFmt, -3),
				func() {
					MultiplexC[int](
						-3,
						nil,
					)(
						nil,
						nil,
						make(chan int),
					)
				},
			)
		},
	)
}
