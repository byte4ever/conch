package conch

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/byte4ever/conch/internal/ratelimit"
)

func TestTee(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()
	oStream, err := DummyTimeGenerator()(ctx)
	require.NoError(t, err)
	require.NotNil(t, oStream)

	const nbSplits = 100
	outputs := MirrorHighThroughput(ctx, oStream, nbSplits)
	require.NotNil(t, outputs)
	require.Len(t, outputs, nbSplits)

	for _, output := range outputs {
		require.NotNil(t, output)
	}

	var wg sync.WaitGroup

	wg.Add(nbSplits)

	for _, output := range outputs {
		go func(stream <-chan time.Time) {
			defer wg.Done()
			countAllAndDelay(stream)
		}(output)
	}

	wg.Wait()
}

func countAll[T any](input <-chan T) {
	var count int
	for range input {
		count++
	}

	fmt.Println(count)
}

func countAllAndDelay(input <-chan time.Time) {
	var count int

	curAvg := time.Since(<-input)

	for t := range input {
		curAvg = (curAvg + time.Since(t)) / 2
		count++
	}

	fmt.Println(count, curAvg)
}

/*
	func TestPriority(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		inputs, output := FairFanIn[struct{}](ctx, 3)

		inputs[0] <- struct{}{}
		inputs[1] <- struct{}{}
		inputs[2] <- struct{}{}

		// Here we are

		<-output
		<-output
		<-output

		// for _, input := range inputs {
		// 	close(input)
		// }
	}
*/
type K int

func (k K) LessThan(other Comparable) bool {
	return k < other.(K)
}

func RandomGen(ctx context.Context) chan K {
	outStream := make(chan K)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return
			case outStream <- K(rand.Intn(10)):
			}
		}
	}()

	return outStream
}

func RandomShuffled(startValue, count int) <-chan K {
	outStream := make(chan K, count)
	values := make([]K, count)

	for i := 0; i < count; i++ {
		values[i] = K(startValue + i)
	}

	for i := 0; i < 4; i++ {
		rand.Shuffle(
			len(values), func(i, j int) {
				values[i], values[j] = values[j], values[i]
			},
		)
	}

	go func() {
		defer close(outStream)

		for _, value := range values {
			outStream <- value
		}
	}()

	return outStream
}

func TestByPriority_Success(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	inCtx, inCancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer inCancel()

	inStream := RandomGen(inCtx)
	out := Prioritize(
		inCtx,
		inStream,
		WithBackPressureDelay(time.Millisecond),
	)

	var cnt int
	for range out {
		cnt++
	}

	t.Logf("got %d items", cnt)
}

func TestByPriority_CloseInput(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	inCtx, inCancel := context.WithCancel(
		context.Background(),
	)
	defer inCancel()

	inStream := RandomGen(inCtx)
	out := Prioritize(
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

	outStream := Prioritize(
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

	outStream := Prioritize(
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

	outStream := Prioritize(
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

func sameGenerator[T any](
	ctx context.Context, id T, wg *sync.WaitGroup,
) chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)
		wg.Wait()

		for {
			select {
			case <-ctx.Done():
				return
			case outStream <- id:
			}
		}
	}()

	return outStream
}

func TestUnfairPriority(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	const (
		nbPrio            = 20
		backPressureDelay = 1 * time.Millisecond
		testDuration      = 10 * time.Second
	)

	var startSignal sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	startSignal.Add(1)

	inStreams := make([]<-chan int, nbPrio)
	for i := 0; i < nbPrio; i++ {
		inStreams[i] = sameGenerator(ctx, i, &startSignal)
	}

	output := UnfairFanIn(ctx, inStreams...)

	stats := make([]uint64, nbPrio)

	startSignal.Done()

	for v := range output {
		time.Sleep(backPressureDelay)
		stats[v]++
	}

	fmt.Println(stats)

	var accum uint64
	for _, v := range stats {
		accum += v
	}

	for _, v := range stats {
		fmt.Println(100.0 * float64(v) / float64(accum))
	}

	fmt.Println(accum)

	fmt.Println(int64(testDuration / backPressureDelay))
}

func TestFairPriority(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	const (
		nbPrio            = 20
		backPressureDelay = 1 * time.Millisecond
		testDuration      = 10 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	inStreams := make([]<-chan int, nbPrio)
	for i := 0; i < nbPrio; i++ {
		inStreams[i] = sameGenerator(ctx, i, &wg)
	}

	output := FairFanIn(ctx, inStreams...)

	stats := make([]uint64, nbPrio)

	wg.Done()

	for v := range output {
		time.Sleep(backPressureDelay)
		stats[v]++
	}

	fmt.Println(stats)

	var accum uint64

	for _, v := range stats {
		accum += v
	}

	for _, v := range stats {
		fmt.Println(100.0 * float64(v) / float64(accum))
	}

	fmt.Println(accum)

	fmt.Println(int64(testDuration / backPressureDelay))
}

func IntSequenceGenerator(
	minValue,
	maxValue int,
) Generator[int] {
	return func(ctx context.Context) (<-chan int, error) {
		output := make(chan int)

		go func() {
			defer close(output)

			for {
				for i := minValue; i <= maxValue; i++ {
					select {
					case <-ctx.Done():
						return
					case output <- i:
					}
				}
			}
		}()

		return output, nil
	}
}

func DummyTimeGenerator() Generator[time.Time] {
	return func(ctx context.Context) (<-chan time.Time, error) {
		output := make(chan time.Time)

		go func() {
			defer close(output)

			for {
				select {
				case <-ctx.Done():
					return
				case output <- time.Now():
				}
			}
		}()

		return output, nil
	}
}

func DelayTimeGenerator(d time.Duration) Generator[time.Time] {
	return func(ctx context.Context) (<-chan time.Time, error) {
		output := make(chan time.Time)

		go func() {
			defer close(output)

			timer := time.NewTimer(d)

			for {
				select {
				case <-ctx.Done():
					if !timer.Stop() {
						<-timer.C
					}

					return

				case <-timer.C:
					select {
					case output <- time.Now():
						timer.Reset(d)
					case <-ctx.Done():
						if !timer.Stop() {
							<-timer.C
						}
						return
					}
				}
			}
		}()

		return output, nil
	}
}

func DummyGenerator() Generator[struct{}] {
	return func(ctx context.Context) (<-chan struct{}, error) {
		output := make(chan struct{})

		go func() {
			defer close(output)

			for {
				select {
				case <-ctx.Done():
					return
				case output <- struct{}{}:
				}
			}
		}()

		return output, nil
	}
}

func TestOpenedValve(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tsStream, _ := DelayTimeGenerator(200 * time.Millisecond)(ctx)
	openValve, closeValve, out := OpenedValve(ctx, tsStream)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for t2 := range out {
			fmt.Println(t2)
		}
	}()

	go func() {
		defer wg.Done()

		var isClosed bool

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				if isClosed {
					fmt.Println("open ---------------------------")
					os.Stdout.Sync()
					openValve()
					isClosed = false
				} else {
					closeValve()
					fmt.Println("closed ---------------------------")
					os.Stdout.Sync()
					isClosed = true
				}
			}
		}
	}()

	wg.Wait()
}

func TestRateLimit(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	timeOutStream, _ := DummyTimeGenerator()(ctx)

	openValve, closeValve, valveOutStream := OpenedValve(ctx, timeOutStream)

	rateLimitedOutStream := RateLimit(
		ctx,
		valveOutStream,
		10,
		// ratelimit.Per(1000*time.Second),
		ratelimit.WithSlack(10/3),
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		var isClosed bool

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				if isClosed {
					fmt.Println("open ---------------------------")
					os.Stdout.Sync()
					openValve()
					isClosed = false
				} else {
					closeValve()
					fmt.Println("closed ---------------------------")
					os.Stdout.Sync()
					isClosed = true
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		for s := range rateLimitedOutStream {
			fmt.Println(s)
		}
	}()

	wg.Wait()
}
