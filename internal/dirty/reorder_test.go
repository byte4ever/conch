package dirty

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch"
)

func randomIndexed(
	t *testing.T,
	ctx context.Context, spread int,
) <-chan dirty.Indexed[int, any] {
	t.Helper()

	outStream := make(chan dirty.Indexed[int, interface{}])

	go func() {
		var cnt int
		defer func() {
			t.Log(cnt, " items sent")
		}()
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case outStream <- dirty.Indexed[int, interface{}]{
				Index: rand.Intn(spread),
			}:
				cnt++
			}
		}
	}()

	return outStream
}

func randomIndexedCount(
	ctx context.Context, count int, spread int,
) <-chan dirty.Indexed[int, interface{}] {
	outStream := make(chan dirty.Indexed[int, interface{}])

	go func() {
		defer close(outStream)
		defer func() {
			fmt.Println(count, "items sent")
		}()

		for i := 0; i < count; i++ {
			v := rand.Intn(spread)
			select {
			case <-ctx.Done():
				return
			case outStream <- dirty.Indexed[int, interface{}]{
				Index: v,
			}:
			}
		}
	}()

	return outStream
}

func TestReorder(t *testing.T) {
	t.Parallel()

	t.Run(
		"streaming closed by context", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			const testDuration = time.Second

			ctx, cancel := context.WithTimeout(
				context.Background(),
				testDuration,
			)
			defer cancel()

			outStream := Reorder(
				ctx,
				randomIndexed(t, ctx, 10000),
			)

			dirty.isClosing(t, outStream, testDuration*2, time.Millisecond)
		},
	)

	t.Run(
		"streaming closed by closing input", func(t *testing.T) {
			t.Parallel()
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			const testDuration = time.Second

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Reorder(
				ctx,
				randomIndexedCount(ctx, 10000, 10000),
			)

			dirty.isClosing(t, outStream, testDuration, time.Millisecond)
		},
	)
}

func TestMe(t *testing.T) {
	t.Parallel()
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var wg sync.WaitGroup

	wg.Add(2)

	c := make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var start sync.WaitGroup

	start.Add(1)

	var sendItems int
	var receivedItems int

	go func() {
		defer wg.Done()
		defer close(c)
		defer func() {
			fmt.Println("send", sendItems, "items")
		}()

		start.Wait()

		for {
			select {
			case <-ctx.Done():
				return
			case c <- struct{}{}:
				sendItems++
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			fmt.Println("received", receivedItems, "items")
		}()

		start.Wait()

		for {
			select {
			case <-ctx.Done():
				return
			case _, more := <-c:
				if !more {
					return
				}
				receivedItems++
			}
		}
	}()

	start.Done()
	wg.Wait()
}

func Test_inOrder(t *testing.T) {
	t.Parallel()
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var h internalHeapInterface[uint64, interface{}]

	const nbValues = 100_000

	pv := make(inOrder[uint64, interface{}], 0, nbValues)

	h = &pv

	values := make([]uint64, nbValues)
	for i := uint64(0); i < nbValues; i++ {
		values[i] = i
	}

	for i := 0; i < 10; i++ {
		rand.Shuffle(
			len(values), func(i, j int) {
				values[i], values[j] = values[j], values[i]
			},
		)
	}

	for _, value := range values {
		heap.Push(
			h,
			dirty.Indexed[uint64, interface{}]{
				Index: value,
			},
		)
	}

	v := uint64(0)

	for h.Len() > 0 {
		item, _ := heap.Pop(h).(dirty.Indexed[uint64, interface{}])
		require.Equal(t, item.Index, v)
		v++
	}

	require.Equal(t, uint64(nbValues), v)
}

func Test_revOrder(t *testing.T) {
	t.Parallel()
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	var h internalHeapInterface[uint64, interface{}]

	const nbValues = 100_000
	pv := make(revOrder[uint64, interface{}], 0, nbValues)

	h = &pv

	values := make([]uint64, nbValues)
	for i := uint64(0); i < nbValues; i++ {
		values[i] = i
	}

	for i := 0; i < 10; i++ {
		rand.Shuffle(
			len(values), func(i, j int) {
				values[i], values[j] = values[j], values[i]
			},
		)
	}

	for _, value := range values {
		heap.Push(
			h,
			dirty.Indexed[uint64, interface{}]{
				Index: value,
			},
		)
	}

	v := uint64(nbValues - 1)

	for h.Len() > 0 {
		item, _ := heap.Pop(h).(dirty.Indexed[uint64, interface{}])
		require.Equal(t, item.Index, v)
		v--
	}

	require.Equal(t, uint64(math.MaxUint64), v)
}
