package conch_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	xxhash2 "github.com/OneOfOne/xxhash"
	"github.com/dgraph-io/ristretto"
	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

type wrapClient struct {
	*retryablehttp.Client
}

func (w *wrapClient) Do(request *http.Request) (*http.Response, error) {
	r, err := retryablehttp.FromRequest(request)
	if err != nil {
		return nil, err
	}

	return w.Client.Do(r)
}

func FibonacciRecursion(n int) int {
	if n <= 1 {
		return n
	}

	return FibonacciRecursion(n-1) + FibonacciRecursion(n-2)
}

type RistrettoWrapper[K conch.Hashable, V any] struct {
	cache *ristretto.Cache
}

func (r *RistrettoWrapper[K, V]) Get(
	_ context.Context,
	key K,
) (V, bool) {
	var val V

	pp, found := r.cache.Get(key.Hash())
	if !found {
		return val, false
	}

	val, _ = pp.(V)

	return val, true
}

func (r *RistrettoWrapper[K, V]) Store(_ context.Context, key K, value V) {
	r.cache.SetWithTTL(key.Hash(), value, 1, 180*time.Second)
}

type Param string

func (V Param) Hash() conch.Key {
	return conch.Key{A: xxhash2.ChecksumString64(string(V))}
}

func TestFullmonty(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	cache, err := ristretto.NewCache(
		&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
			// Metrics:     true,
			KeyToHash: func(
				key any,
			) (uint64, uint64) {
				pk, _ := key.(conch.Key)
				return pk.A, pk.B
			},
		},
	)

	if err != nil {
		panic(err)
	}

	/*wg.Add(1)

	go func() {
		defer wg.Done()

		const metricsFrequency = 2 * time.Second
		timer := time.NewTimer(metricsFrequency)
		m := cache.Metrics
		for {
			select {
			case <-timer.C:
				fmt.Println(
					"----------------- metrics",
					m.KeysAdded()-m.KeysEvicted(),
					m.Ratio(),
				)
				timer.Reset(metricsFrequency)
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			}
		}
	}()
	*/
	wrappedCache := &RistrettoWrapper[Param, *Resp]{cache: cache}

	// cli := &wrapClient{Client: retryablehttp.NewClient()}

	// c := InfoClient{
	// 	HTTPDoerWithContext: cli,
	// }

	const concurrencyFactor = 32

	requester := conch.RequesterC(
		ctx,
		&wg,
		conch.CacheReadInterceptorPoolC[Param, *Resp](
			concurrencyFactor,
			wrappedCache,
			conch.DedupC(
				conch.PoolC(
					concurrencyFactor,
					conch.CacheInterceptorC[Param, *Resp](
						wrappedCache,
						conch.RequestConsumerC(
							// c.Get,
							func() func(
								context.Context,
								Param,
							) (*Resp, error) {
								rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
								return func(
									ctx context.Context,
									p Param,
								) (*Resp, error) {
									fmt.Println("---------------- GET", p)
									time.Sleep(
										time.Duration(
											rnd.Int63n(
												5000,
											),
										) * time.Microsecond,
									)
									return &Resp{
										Population: 100,
									}, nil
								}
							}(),
						),
					),
				),
			),
		),
	)

	const INSEECount = 2000

	INSEECodesSet := GetRandomINSEECodes(INSEECount)

	var cntSend atomic.Uint64
	var cntReceived atomic.Uint64

	callStream := make(chan struct{})
	conch.PoolC(
		concurrencyFactor+1,
		conch.ConsumerC(
			func() func(context.Context, struct{}) {
				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				return func(ctx context.Context, t struct{}) {
					// s := time.Now()
					p := rnd.Intn(INSEECount)

					cntSend.Add(1)
					_, _ = requester(ctx, Param(INSEECodesSet[p]))
					cntReceived.Add(1)
					// d := time.Since(s)
					// if err != nil {
					// 	fmt.Println(d, err)
					//
					// } else {
					// 	fmt.Println(d, r.Population)
					// }
					// time.Sleep(time.Second)
				}
			}(),
		),
	)(ctx, &wg, callStream)

	for i := 0; i < 1_000_000; i++ {
		callStream <- struct{}{}
		// time.Sleep(time.Second)
	}

	cancel()
	wg.Wait()
	cache.Close()

	fmt.Println(cntSend.Load())
	fmt.Println(cntReceived.Load())
}
