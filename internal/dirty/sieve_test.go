package dirty

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type sieveTime struct {
	id       int
	createAt time.Time
}

func index(s *sieveTime) int {
	return s.id
}

func sieveTimeGenerator(
	ctx context.Context,
	maxId int,
) <-chan *sieveTime {
	outStream := make(chan *sieveTime)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case outStream <- &sieveTime{
				id:       rand.Intn(maxId),
				createAt: time.Now(),
			}:
			}
		}
	}()

	return outStream
}

func countSieveTime(i int, input <-chan *sieveTime) {
	var count int

	t, more := <-input
	if !more {
		return
	}

	count++

	curAvg := time.Since(t.createAt)

	for t := range input {
		curAvg = (curAvg + time.Since(t.createAt)) / 2
		count++
	}

	fmt.Printf("--- #%d %d %v\n", i, count, curAvg.Microseconds())
}

func TestSieve(t *testing.T) {
	const nbOutputs = 80

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	outSieveTimeStream := sieveTimeGenerator(ctx, nbOutputs)

	outStreams := Sieve(ctx, index, nbOutputs, outSieveTimeStream)

	var wg sync.WaitGroup

	wg.Add(nbOutputs)

	for i, output := range outStreams {
		go func(stream <-chan *sieveTime, idx int) {
			defer wg.Done()
			countSieveTime(idx, stream)
		}(output, i)
	}

	wg.Wait()
}

func Test_recBuildTeeTreeIdx(t *testing.T) {
	ctx := context.Background()
	inStream := make(chan sieveItem[int])

	const nbOut = 31
	out := make([]<-chan sieveItem[int], nbOut)
	recBuildTeeTreeIdx(
		ctx,
		0,
		inStream,
		0,
		nbOut,
		out,
	)
	fmt.Println(out)
}

/*
Func TestSieve2(t *testing.T) {
	const nbOutputs = 80

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	outSieveTimeStream := sieveTimeGenerator(ctx, nbOutputs)

	outStreams := SieveTree(ctx, index, nbOutputs, outSieveTimeStream)

	var wg sync.WaitGroup

	wg.Add(nbOutputs)

	for i, output := range outStreams {
		go func(stream <-chan *sieveTime, idx int) {
			defer wg.Done()
			countSieveTime(idx, stream)
		}(output, i)
	}

	wg.Wait()
}
*/

func TestSieveShit(t *testing.T) {
	const nbOutputs = 80

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	outSieveTimeStream := sieveTimeGenerator(ctx, nbOutputs)

	outStreams := SieveTree(ctx, index, nbOutputs, outSieveTimeStream)

	fmt.Println(outStreams)

	var wg sync.WaitGroup

	wg.Add(nbOutputs)

	for i, output := range outStreams {
		go func(stream <-chan *sieveTime, idx int) {
			defer wg.Done()
			countSieveTime(idx, stream)
		}(output, i)
	}

	wg.Wait()
}

/*
Func choice[T any](inStream <-chan T, idx int) (_, _ <-chan T) {
	outStream1 := make(chan T, 1)
	outStream2 := make(chan T, 1)

	go func() {
		defer closeMe(outStream1)
		defer closeMe(outStream2)

		for val := range inStream {
			var out1, out2 = outStream1, outStream2

			for i := 0; i < 2; i++ {
				select {
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()

	return outStream1, outStream2
}
*/

/*Func Test_recBuildTeeTreeIdx(t *testing.T) {
	ctx := context.Background()
	inStream := make(chan sieveItem[int])

	const nbOut = 31
	out := make([]<-chan sieveItem[int], nbOut)
	recBuildTeeTreeIdx(
		ctx,
		0,
		inStream,
		0,
		nbOut,
		out,
	)
	fmt.Println(out)
}
*/

func TestSieve2(t *testing.T) {
	const nbOutputs = 10

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testStream := make(chan *sieveTime)

	outStreams := SieveTree(ctx, index, nbOutputs, testStream)

	var wg sync.WaitGroup

	var start sync.WaitGroup

	start.Add(1)
	wg.Add(nbOutputs + 1)

	go func() {
		defer wg.Done()
		defer close(testStream)
		start.Wait()
		wc := time.Now()
		const cc = 600_000
		for i := 0; i < cc; i++ {
			// if i%1000 == 0 {
			// 	fmt.Println(i)
			// }
			testStream <- &sieveTime{
				id:       rand.Intn(nbOutputs),
				createAt: time.Now(),
			}
		}
		d := time.Since(wc)
		rate := float64(cc) / d.Seconds()
		fmt.Println("pushed in ", d, rate)
	}()

	for i, output := range outStreams {
		go func(stream <-chan *sieveTime, idx int) {
			defer wg.Done()
			start.Wait()
			if idx == 7 {
				countSieveTime(
					idx, RateLimit(
						ctx,
						stream,
						1000,
					),
				)
				return
			}

			countSieveTime(idx, stream)
		}(output, i)
	}

	start.Done()
	wg.Wait()
}
