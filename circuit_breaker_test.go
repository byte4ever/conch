package conch

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

var (
	// ErrBadass represent an error where ....
	ErrBadass = errors.New("badass")

	// ErrUnavailable represent an error where....
	ErrUnavailable = errors.New("unavailable")
)

func TestBreaker(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	const count = 300

	testDuration := 2 * time.Second

	for x := 0; x < count; x++ {
		fmt.Println("test: ", x)

		var wg sync.WaitGroup

		ctx, cancel := context.WithTimeout(context.Background(), testDuration)
		defer cancel()

		const concurrency = 16

		requester := RequesterC(
			ctx,
			&wg,
			BreakerC(
				4,
				3,
				20*time.Millisecond,
				ErrUnavailable,
				RequestConsumerPoolC(
					func(ctx context.Context, p2 int) (int, error) {
						if rand.Float64() < 0.1 {
							return 0, ErrBadass
						}
						//time.Sleep(5 * time.Millisecond)
						return p2, nil
					},
					concurrency,
				),
			),
		)

		var (
			wgStart sync.WaitGroup
		)

		//counters := make([]int, concurrency)

		wgStart.Add(1)

		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func(
				idx int,
				wgStart *sync.WaitGroup,
				wg *sync.WaitGroup,
			) {
				//defer fmt.Println("requester ", idx, "done")
				defer wg.Done()

				var p int

				wgStart.Wait()
				//fmt.Println("requester ", idx, "start")

				for {
					select {
					case <-ctx.Done():
						return

					default:
						_, _ = requester(ctx, p)
						//counters[idx]++
						p++
					}
				}
			}(
				i,
				&wgStart,
				&wg,
			)
		}

		wgStart.Done()
		fmt.Println("requesters started")

		require.Eventually(
			t,
			func() bool {
				wg.Wait()
				return true
			},
			testDuration+time.Second,
			10*time.Millisecond,
		)

		fmt.Println("all done")

		//fmt.Println(counters)

		//var sum int
		//for _, counter := range counters {
		//	sum += counter
		//}
		//
		//fmt.Println(sum)
	}

}
