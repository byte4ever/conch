package dirty

import (
	"context"
	"math/rand"
	"sort"
	"sync"

	"github.com/byte4ever/conch"
)

type rouletteItem struct {
	k   float64
	idx int
}

type roulette []rouletteItem

func newRoulette(
	weights []float64,
) (roulette, error) {
	var sum float64

	// compute sum of weights
	for _, w := range weights {
		sum += w
	}

	result := make(roulette, len(weights))

	// normalize weights
	for i, w := range weights {
		result[i] = rouletteItem{
			k:   w / sum,
			idx: i,
		}
	}

	sort.Slice(
		result, func(i, j int) bool {
			return result[i].k < result[j].k
		},
	)

	for i := 1; i < len(weights); i++ {
		result[i].k += result[i-1].k
	}

	return result, nil
}

func (r roulette) Pick(rnd *rand.Rand) int {
	k := rnd.Float64()

	return sort.Search(
		len(r), func(i int) bool {
			return r[i].k >= k
		},
	)
}

func SplitRoulette[T any](
	ctx context.Context,
	rnd *rand.Rand,
	weights []float64,
	inStream <-chan T,
) []<-chan T {
	roulette, err := newRoulette(weights)

	if err != nil {
		panic(err)
	}

	s := Transform(
		ctx,
		func(_ context.Context, v T) dirty.IndexedInteger[int, T] {
			return dirty.IndexedInteger[int, T]{
				Index:   roulette.Pick(rnd),
				Payload: v,
			}
		}, inStream,
	)

	return Spread(ctx, s, len(weights))
}

func SplitRouletteC[T any](
	rnd *rand.Rand,
	weights []float64,
	out ...dirty.ChainFunc[T],
) dirty.ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		streams := SplitRoulette(ctx, rnd, weights, inStream)
		for idx, s := range streams {
			out[idx](ctx, wg, s)
		}
	}
}
