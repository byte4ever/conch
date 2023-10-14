package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

var (
	ErrMainError = errors.New("main error")
)

func doit(_ context.Context, p int) (int, error) {
	if p%2 == 0 {
		return 0, ErrMainError
	}

	return p*17 + 11, nil
}

func main() {
	// f, err := os.Create("mem.prof")
	// if err != nil {
	// 	log.Fatal("could not create memory profile: ", err)
	// }
	// defer f.Close() // error handling omitted for example
	// runtime.GC()    // get up-to-date statistics

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requester := conch.RequesterC(
		ctx,
		&wg,
		conch.RequestConsumerC(
			doit,
		),
	)

	var p time.Duration

	const ConcurrentRequests = 10_000_000

	for i := 0; i < ConcurrentRequests; i++ {
		s := time.Now()
		r, err := requester(context.Background(), i)
		p += time.Since(s)
		if i%2 == 0 {
			if !errors.Is(err, ErrMainError) {
				panic("shit")
			}

			continue
		}

		if i*17+11 != r {
			panic("shit 2")
		}
	}

	cancel()

	wg.Wait()

	fmt.Println("time:", p/time.Duration(ConcurrentRequests))

	// if err := pprof.WriteHeapProfile(f); err != nil {
	// 	log.Fatal("could not write memory profile: ", err)
	// }
}
