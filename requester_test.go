package conch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"
)

var errF = errors.New("invalid value")

func f(_ context.Context, v int) (int, error) {
	time.Sleep(2 * time.Millisecond)

	return v, nil
}

func TestRequester(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const prioSpread = 10

	prioRequesters := make(
		[]func(
			context.Context,
			int,
		) (
			int,
			error,
		),
		prioSpread,
	)

	streams := make(
		[]<-chan Request[int, int],
		prioSpread,
	)

	for i := 0; i < prioSpread; i++ {
		// var s <-chan Request[int, int]
		// prioRequesters[i], s = Requester[int, int](ctx)
		// streams[i] = Buffer(ctx, s, 100000)
		prioRequesters[i], streams[i] = Requester[int, int](ctx)
	}

	unfairStream := UnfairFanIn(ctx, streams...)

	watchStream := unfairStream
	// watchStream := Transform(
	// ctx,
	// unfairStream,
	// func(
	// 	ctx context.Context,
	// 	v Request[int, int],
	// ) Request[int, int] {
	// 	fmt.Println(v.P)
	// 	return v
	// },
	// )

	SpawnRequestProcessorsPool(ctx, watchStream, f, 4, "proc")

	const parallelism = prioSpread
	var wg, start sync.WaitGroup

	wg.Add(parallelism)
	start.Add(1)

	var totCnt atomic.Int64

	for i := 0; i < parallelism; i++ {
		go func(i int) {
			defer wg.Done()

			var (
				cnt int64
				md  time.Duration
			)

			first := true

			start.Wait()
			end := time.Now().Add(20 * time.Second)
			for time.Now().Before(end) {
				s := time.Now()
				_, _ = prioRequesters[i](ctx, i)
				d := time.Since(s)

				if first {
					md = d
				} else {
					md = (md + d) / 2
				}

				cnt++

				first = false
				// time.Sleep(4 * time.Millisecond)
			}

			fmt.Println(i, cnt, md)
			totCnt.Add(cnt)
		}(i)
	}

	start.Done()
	wg.Wait()
	fmt.Println(totCnt.Load())
}
