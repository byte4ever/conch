package conch

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch/cbreaker"
	"github.com/byte4ever/conch/contrib/cache"
	"github.com/byte4ever/conch/domain"
	"github.com/byte4ever/conch/retrier"
)

type Elem struct {
	param cache.HashableUint64
	value uint64
	err   error
}

func (f *Elem) GetParam() cache.HashableUint64 {
	return f.param
}

func (f *Elem) SetValue(r uint64) {
	f.value = r
	f.err = nil
}
func (f *Elem) SetError(e error) {
	f.value = 0
	f.err = e
}

func (f *Elem) GetError() error {
	return f.err
}

func (f *Elem) GetValue() uint64 {
	return f.value
}

func RecFibonacci(n uint64) uint64 {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	return RecFibonacci(n-1) + RecFibonacci(n-2)
}

func FiboProcessor(
	_ context.Context,
	elem *Elem,
) {

	if rand.Float64() < 0.4 {
		elem.SetError(retrier.WrapToRetryableError(ErrMocked))
		return
	}

	param := elem.GetParam()
	if param == 0 || param == 1 {
		elem.SetValue(1)
		return
	}

	fmt.Println("------- COMPUTING ", uint64(param), " -------")
	result := RecFibonacci(uint64(param))
	elem.SetValue(result)
}

var _ domain.Processable[cache.HashableUint64, uint64] = &Elem{}

var _ domain.ProcessorFunc[*Elem, cache.HashableUint64, uint64] = FiboProcessor

func TestBolos(t *testing.T) {

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
		KeyToHash:   cache.RistrettoKeyToHash,
	})

	require.NoError(t, err)
	require.NotNil(t, ristrettoCache)

	wrappedCache := cache.WrapRistretto[cache.HashableUint64, uint64](
		ristrettoCache,
		600*time.Second,
	)
	require.NotNil(t, wrappedCache)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewPool[Elem](1000)

	inStream := make(chan *Elem)

	cb, err := cbreaker.NewEngine()

	require.NoError(t, err)
	require.NotNil(t, cb)

	rt, err := retrier.New(
		retrier.WithMaxRetry(3),
	)

	f := cache.
		WrapProcessorFunc(
			wrappedCache,
			cbreaker.
				WrapProcessorFunc(cb,
					retrier.WrapProcessorFunc(rt,
						FiboProcessor,
					),
				),
		)

	_ = cbreaker.
		WrapProcessorFunc(cb,
			retrier.WrapProcessorFunc(rt,
				FiboProcessor,
			),
		)

	MultiplexC(
		4,
		ProcessorsC(
			f,
			// MultiplexC(20,
			SinksC[Elem](
				pool,
				func(_ context.Context, v *Elem) {
					fmt.Println(
						time.Now().Format(time.RFC3339Nano),
						v.GetParam(),
						v.GetValue(),
						v.GetError(),
					)
				},
			),
			// ),
		),
	)(ctx, &wg, inStream)

	wg.Add(1)

	go func() {
		defer func() {
			close(inStream)
			wg.Done()
		}()

		const maxFibIdx = 47
		values := make([]cache.HashableUint64, 0, maxFibIdx)

		for i := range maxFibIdx {
			values = append(values, cache.HashableUint64(i))
		}

		for {
			rand.Shuffle(len(values), func(i, j int) {
				values[i], values[j] = values[j], values[i]
			})

			for _, i := range values {
				f := pool.Get()
				f.param = i
				inStream <- f
			}
		}
	}()

	wg.Wait()
}
