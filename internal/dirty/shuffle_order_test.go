package dirty

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/byte4ever/conch"
)

func genStream[T any](gen func(int) T, count int) <-chan T {
	outStream := make(chan T, count)

	for i := 0; i < count; i++ {
		outStream <- gen(i)
	}

	close(outStream)

	return outStream
}

func notInOrder[T dirty.Ordered](t *testing.T, vals []T) {
	if len(vals) == 0 {
		t.Fatalf("not enough values")
	}

	disorderCount := 0

	for i := 1; i < len(vals); i++ {
		if vals[i-1] > vals[i] {
			disorderCount++
		}
	}

	require.NotZerof(t, disorderCount, "in order")
}

func nodDup[T comparable](t *testing.T, vals []T) {
	if len(vals) <= 1 {
		return
	}

	deDup := make(map[T]int, len(vals))

	for i := 1; i < len(vals); i++ {
		deDup[vals[i]]++
	}

	if len(deDup) != len(vals) {
		dupVals := make([]any, 0, len(vals))
		dupFmt := ""

		for k, v := range deDup {
			if v > 1 {
				dupVals = append(dupVals, k)
				dupFmt += " %v"
			}
		}

		if len(dupVals) > 0 {
			require.Failf(t, "duplicate values found", dupFmt, dupVals...)
		}
	}

}

func TestShuffleOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1 := genStream(func(i int) int { return i }, 100)

	rnd := rand.New(rand.NewSource(0))

	out := ShuffleOrder(ctx, 17, rnd, s1)

	vals := <-ToList(ctx, out, 100)

	notInOrder(t, vals)
	nodDup(t, vals)
	require.Len(t, vals, 100)
	sort.Ints(vals)
	fmt.Println(vals)
	require.Equal(t, 0, vals[0])
	require.Equal(t, 99, vals[len(vals)-1])
}
