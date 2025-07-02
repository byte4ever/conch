package tests

import (
	"github.com/byte4ever/conch/domain"
)

type PP uint64

func (p PP) Hash() domain.Key {
	return domain.Key{
		A: uint64(p),
		B: uint64(p),
	}
}

/*
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
		60*time.Second,
	)
	require.NotNil(wrappedCache)

	defer goleak.VerifyNone(suite.T(), goleak.IgnoreCurrent())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var group sync.WaitGroup

	var beforeDedupCounter atomic.Uint64
	var beforeCacheCounter atomic.Uint64
	var afterCacheCounter atomic.Uint64

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		var beforeCacheCounterP, afterCacheCounterP, beforeDedupCounterP uint64

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				beforeDedupCounter := beforeDedupCounter.Load()
				beforeCacheCounter := beforeCacheCounter.Load()
				afterCacheCounter := afterCacheCounter.Load()
				fmt.Printf(
					"%8d %8d %8d %8d %8d %8d\n",
					beforeDedupCounter,
					beforeDedupCounter-beforeDedupCounterP,
					beforeCacheCounter,
					beforeCacheCounter-beforeCacheCounterP,
					afterCacheCounter,
					afterCacheCounter-afterCacheCounterP,
				)
				beforeDedupCounterP = beforeDedupCounter
				beforeCacheCounterP = beforeCacheCounter
				afterCacheCounterP = afterCacheCounter
			}
		}
	}()

	const nbWorkers = 20

	requester := conch.RequesterC(
		ctx,
		&group,
		conch.CountsC(
			&beforeDedupCounter,
			conch.DeduplicatorsC(
				conch.CountsC(
					&beforeCacheCounter,
					conch.CacheReadInterceptorsC[PP, uint64](
						wrappedCache,
						conch.CountsC(
							&afterCacheCounter,
							conch.MultiplexC[conch.Request[PP, uint64]](
								nbWorkers,
								conch.CacheWriteInterceptorsC[PP, uint64](
									wrappedCache,
									conch.RequestConsumersC[PP, uint64](
										func(
											ctx context.Context,
											id int,
											p PP,
										) (
											uint64,
											error,
										) {
											if p > 44 {
												return 0, ErrToManyIteration
											}
											result := FibonacciRecursion(
												uint64(p),
											)
											return result,
												nil
										},
									),
								),
							),
						),
					),
				),
			),
		),
	)

	const (
		nbRequestPerRequester = 100_000
		nbRequesterThreads    = 10
	)

	var runGroup sync.WaitGroup
	runGroup.Add(nbRequesterThreads)
	for i := 0; i < nbRequesterThreads; i++ {
		go func(idx int) {
			defer runGroup.Done()
			time.Sleep(
				time.Duration(intervalInt63nRand(500, 10_000)) *
					time.Millisecond,
			)
			for i := 0; i < nbRequestPerRequester; i++ {
				require.Eventually(func() bool {
					n := intervalInt63nRand(1, 10)
					_, _ = requester(context.Background(), PP(n))
					return true
				}, 90*time.Second, 10*time.Millisecond)
			}
		}(i)
	}

	runGroup.Wait()

	cancel()
	group.Wait()

	fmt.Println(beforeDedupCounter.Load())
	fmt.Println(beforeCacheCounter.Load())
	fmt.Println(afterCacheCounter.Load())
}

// ErrToManyIteration indicates that the maximum iteration count was exceeded.
var ErrToManyIteration = errors.New("to many iteration")

func FibonacciRecursion(n uint64) uint64 {
	if n <= 1 {
		return n
	}
	return FibonacciRecursion(n-1) + FibonacciRecursion(n-2)
}

func intervalInt63nRand(a, b int64) int64 {
	return a + rand.Int63n(b-a)
}
*/
