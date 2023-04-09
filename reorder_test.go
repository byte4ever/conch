package conch

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func randomIndexed(ctx context.Context) <-chan Indexed[int, interface{}] {
	outStream := make(chan Indexed[int, interface{}])

	go func() {
		var cnt int
		defer func() {
			fmt.Println("sent ", cnt, " items")
		}()
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case outStream <- Indexed[int, interface{}]{
				Index: rand.Intn(100),
			}:
				cnt++
			}
		}
	}()

	return outStream
}

func randomIndexedCount(
	ctx context.Context,
	count int,
) <-chan Indexed[int, interface{}] {
	outStream := make(chan Indexed[int, interface{}])

	go func() {
		defer close(outStream)
		defer func() {
			fmt.Println("bye sender")
		}()

		for i := 0; i < count; i++ {
			v := rand.Intn(10)
			select {
			case <-ctx.Done():
				return
			case outStream <- Indexed[int, interface{}]{
				Index: v,
			}:
			}
		}
	}()

	return outStream
}

func TestReorder(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	t.Parallel()

	t.Run(
		"streaming closed by context", func(t *testing.T) {
			const testDuration = 60 * time.Second

			t.Parallel()

			ctx, cancel := context.WithTimeout(
				context.Background(),
				testDuration,
			)
			defer cancel()

			// ctx := context.Background()
			//
			// st := make(chan Indexed[uint64, interface{}])
			//
			// go func() {
			// 	defer close(st)
			//
			// 	for i := uint64(0); i < 100000; i++ {
			// 		st <- Indexed[uint64, interface{}]{
			// 			Index: i,
			// 		}
			// 	}
			// }()

			outStream := Reorder(
				ctx,
				// st,
				randomIndexed(ctx),
				// WithBufferSize(5),
				// WithOrderReversed(),
			)

			// require.Eventually(
			// 	t, func() bool {
			var cnt int
			for range outStream {
				cnt++
			}

			// for o := range outStream {
			// 	fmt.Println(cnt, o.Index)
			// 	cnt++
			// }

			// require.NotZero(t, cnt)
			fmt.Printf("received  #%d items\n", cnt)
			fmt.Printf(
				"rate  %v iter/s\n",
				float64(cnt)/testDuration.Seconds(),
			)
			fmt.Printf(
				"iter duration %v\n",
				testDuration/time.Duration(cnt),
			)

			// return true
			// },
			// testDuration+1*time.Millisecond,
			// 10*time.Millisecond,
			// )

		},
	)

	t.Run(
		"streaming closed by closing input", func(t *testing.T) {
			const testDuration = time.Second

			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			outStream := Reorder(
				ctx,
				randomIndexedCount(ctx, 20),
			)

			require.Eventually(
				t, func() bool {
					var cnt int

					for range outStream {
						cnt++
						fmt.Println(cnt)
					}

					require.NotZero(t, cnt)
					t.Logf("received  #%d items\n", cnt)
					return true
				},
				testDuration+1*time.Millisecond,
				time.Millisecond,
			)
		},
	)
}

/*
func TestByPriority_CloseInput(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	inCtx, inCancel := context.WithCancel(
		context.Background(),
	)
	defer inCancel()

	inStream := RandomGen(inCtx)
	out := Reorder(
		inCtx,
		inStream,
		WithBackPressureDelay(time.Millisecond),
	)

	inCancel()

	var cnt int
	for range out {
		cnt++
	}

	require.Zerof(
		t,
		cnt,
		"got %d items, must get 0 if input stream is closed before reading",
		cnt,
	)
}

func TestByPriority_Order(t *testing.T) {
	const nbValues = 10

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	shuffled := RandomShuffled(0, nbValues)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outStream := Reorder(
		ctx,
		shuffled,
		WithBackPressureDelay(10*time.Millisecond),
		WithBufferSize(nbValues),
		WithFlushWhenDown(),
	)

	cnt := 0

	for k := range outStream {
		require.Equal(t, k, K(cnt))
		cnt++
	}

	require.Equal(t, nbValues, cnt)
}

func TestByPriority_QuickCancel(t *testing.T) {
	const nbValues = 10

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	shuffled := RandomShuffled(0, nbValues)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outStream := Reorder(
		ctx,
		shuffled,
		WithBackPressureDelay(1*time.Second),
		WithBufferSize(nbValues),
	)

	cancel()

	for range outStream {
		require.FailNow(
			t,
			"should not get any items",
		)
		// drain outstream
	}
}

func TestByPriority_PartialOrder(t *testing.T) {
	const nbValues = 100

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	shuffled := RandomShuffled(0, nbValues)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outStream := Reorder(
		ctx,
		shuffled,
		WithBackPressureDelay(1000*time.Millisecond),
		WithBufferSize(nbValues/7),
		WithFlushWhenDown(),
	)

	values := make([]int, nbValues)

	cnt := 0

	for k := range outStream {
		values[cnt] = int(k)
		cnt++
	}

	if !assert.Equal(t, nbValues, cnt) {
		// print values
		for i, v := range values {
			t.Logf("got %5d %5d", i, v)
		}
	}

	valKeep := make([]int, nbValues)
	copy(valKeep, values)
	sort.Ints(values)
	require.Equal(t, 0, values[0])
	require.Equal(t, nbValues-1, values[len(values)-1])

	for i, value := range values {
		// fmt.Println(i, value)
		if !assert.Equal(t, i, value) {
			for i, value := range values {
				fmt.Println(i, value, valKeep[i])
			}
			break
		}
	}
}
*/

func Test_inOrder_Len(t *testing.T) {
	var h heap.Interface

	pv := make(inOrder[uint64, interface{}], 0, 100)

	h = &pv

	values := make([]uint64, 100)
	for i := uint64(0); i < 100; i++ {
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
			Indexed[uint64, interface{}]{
				Index: value,
			},
		)
	}
	for i := uint64(0); i < 100; i++ {
		item, _ := heap.Pop(h).(Indexed[uint64, interface{}])
		fmt.Println(item.Index, i)
	}

}
