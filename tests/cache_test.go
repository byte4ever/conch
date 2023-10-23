package tests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/byte4ever/conch"
	"github.com/byte4ever/conch/contrib/cache"
)

type PP uint64

func (p PP) Hash() conch.Key {
	return conch.Key{
		A: uint64(p),
		B: uint64(p),
	}
}

func (suite *ConchTestSuite) TestCache() {
	require := suite.Require()

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		KeyToHash:   cache.RistrettoKeyToHash,
	})

	require.NoError(err)
	require.NotNil(ristrettoCache)

	wrappedCache := cache.WrapRistretto[PP, uint64](
		ristrettoCache,
		15*time.Second,
	)
	require.NotNil(wrappedCache)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var group sync.WaitGroup

	var beforeCacheCounter atomic.Uint64
	var afterCacheCounter atomic.Uint64

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		var beforeCacheCounterP, afterCacheCounterP uint64

		for {
			select {
			case <-ticker.C:
				beforeCacheCounter := beforeCacheCounter.Load()
				afterCacheCounter := afterCacheCounter.Load()
				fmt.Printf(
					"%8d %8d %8d %8d\n",
					beforeCacheCounter,
					beforeCacheCounter-beforeCacheCounterP,
					afterCacheCounter,
					afterCacheCounter-afterCacheCounterP,
				)
				beforeCacheCounterP = beforeCacheCounter
				afterCacheCounterP = afterCacheCounter
			}
		}
	}()

	requester := conch.RequesterC(
		ctx,
		&group,
		conch.CountsC(
			&beforeCacheCounter,
			conch.CacheReadInterceptorsC(
				wrappedCache,
				conch.CountsC(
					&afterCacheCounter,
					conch.MultiplexC(
						32,
						conch.CacheWriteInterceptorsC(
							wrappedCache,
							conch.RequestConsumersC(
								func(
									ctx context.Context,
									p PP,
								) (
									uint64,
									error,
								) {
									if p > 43 {
										return 0, ErrToManyIteration
									}
									return FibonacciRecursion(uint64(p)), nil
								}),
						),
					),
				),
			),
		),
	)

	var runGroup sync.WaitGroup
	runGroup.Add(30)
	for i := 0; i < 30; i++ {
		go func() {
			defer runGroup.Done()
			for i := 0; i < 1_000_000; i++ {
				n := rand.Int63n(44)

				//s := time.Now()
				_, err := requester(context.Background(), PP(n))
				//d := time.Since(s)

				//if n >= 40 && d > time.Millisecond {
				//fmt.Printf(
				//	"%2d %12.6f\n",
				//	n,
				//	d.Seconds(),
				//)
				//}

				require.NoError(err)
				//time.Sleep(time.Millisecond)
			}
		}()
	}

	//here am i trying to check cache.

	runGroup.Wait()

	cancel()
	group.Wait()

	fmt.Println(beforeCacheCounter.Load())
	fmt.Println(afterCacheCounter.Load())
}

// ErrToManyIteration represent an error where ....
var ErrToManyIteration = errors.New("to many iteration")

func FibonacciRecursion(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return FibonacciRecursion(n-1) + FibonacciRecursion(n-2)
}
