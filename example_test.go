package conch_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/byte4ever/conch"
)

// Example demonstrates a minimal pipeline using conch.
func Example() {
	ctx := context.Background()

	// Create an injector to feed data into the pipeline
	injectFunc, inStream := conch.Injector[int](ctx)

	// Start a goroutine to read from the stream
	go func() {
		for value := range inStream {
			fmt.Println(value)
		}
	}()

	// Inject some data into the pipeline
	for i := 0; i < 3; i++ {
		_ = injectFunc(ctx, i)
	}

	// Give time for processing
	time.Sleep(10 * time.Millisecond)
	// Output:
	// 0
	// 1
	// 2
}

// ExampleBatch demonstrates batching elements from a stream.
func ExampleBatch() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Create input channel
	inputChan := make(chan int, 10)

	// Create batch stream with size 2
	batchStream := conch.Batch(ctx, &wg, 2, inputChan)

	// Collect batches
	var batches [][]int
	go func() {
		for batch := range batchStream {
			batches = append(batches, batch)
		}
	}()

	// Send data synchronously
	inputChan <- 1
	inputChan <- 2
	inputChan <- 3
	inputChan <- 4
	close(inputChan)

	wg.Wait()

	// Print batches
	for _, batch := range batches {
		fmt.Println(batch)
	}
	// Output:
	// [1 2]
	// [3 4]
}

// ExampleMultiplex demonstrates multiplexing streams for fan-in operation.
func ExampleMultiplex() {
	ctx := context.Background()

	// Create input channels
	chan1 := make(chan int)
	chan2 := make(chan int)

	// Multiplex to single output (fan-in)
	outStreams := conch.Multiplex(ctx, 1, chan1, chan2)
	outStream := outStreams[0]

	// Start goroutines to send data
	go func() {
		defer close(chan1)
		chan1 <- 1
		chan1 <- 2
	}()

	go func() {
		defer close(chan2)
		chan2 <- 3
		chan2 <- 4
	}()

	// Collect and sort output for deterministic results
	var results []int
	for val := range outStream {
		results = append(results, val)
	}

	// Sort for deterministic output
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j] > results[j+1] {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	for _, val := range results {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
}

// ExampleUnbatch demonstrates unbatching slices into individual elements.
func ExampleUnbatch() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Create input channel for slices
	inputChan := make(chan []int)

	// Create unbatch stream
	unbatchStream := conch.Unbatch(ctx, &wg, inputChan)

	// Collect elements
	var elements []int
	go func() {
		for elem := range unbatchStream {
			elements = append(elements, elem)
		}
	}()

	// Send batches
	inputChan <- []int{1, 2, 3}
	inputChan <- []int{4, 5}
	close(inputChan)

	wg.Wait()

	// Print elements
	for _, elem := range elements {
		fmt.Println(elem)
	}
	// Output:
	// 1
	// 2
	// 3
	// 4
	// 5
}

// ExampleInjector demonstrates creating an injector for feeding data into a pipeline.
func ExampleInjector() {
	ctx := context.Background()

	// Create an injector
	injectFunc, stream := conch.Injector[string](ctx)

	// Start a goroutine to read from the stream
	go func() {
		for value := range stream {
			fmt.Println(value)
		}
	}()

	// Inject some data
	_ = injectFunc(ctx, "hello")
	_ = injectFunc(ctx, "world")
	_ = injectFunc(ctx, "conch")

	// Give time for processing
	time.Sleep(10 * time.Millisecond)
	// Output:
	// hello
	// world
	// conch
}

// ExampleSink demonstrates consuming data from a stream.
func ExampleSink() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Create input channel
	inputChan := make(chan *int)

	// Create a simple pool receiver that does nothing
	poolReceiver := &simplePoolReceiver[int]{}

	// Create sink to consume data
	conch.Sink(ctx, &wg, poolReceiver, func(ctx context.Context, item *int) {
		fmt.Println(*item)
	}, inputChan)

	// Send data
	go func() {
		defer close(inputChan)
		for i := 1; i <= 3; i++ {
			value := i
			inputChan <- &value
		}
	}()

	wg.Wait()
	// Output:
	// 1
	// 2
	// 3
}

// simplePoolReceiver implements a no-op pool receiver for examples
type simplePoolReceiver[T any] struct{}

func (p *simplePoolReceiver[T]) PutBack(item *T) {
	// Do nothing - just for example purposes
}

// ExampleCount demonstrates counting elements flowing through a stream.
func ExampleCount() {
	ctx := context.Background()
	counter := &atomic.Uint64{}

	// Create input channel
	inputChan := make(chan int)

	// Create count stream
	countStream := conch.Count(ctx, counter, inputChan)

	// Start goroutine to read from count stream
	go func() {
		for val := range countStream {
			fmt.Printf("Value: %d\n", val)
		}
	}()

	// Send data
	go func() {
		defer close(inputChan)
		for i := 1; i <= 3; i++ {
			inputChan <- i
		}
	}()

	// Wait a bit for processing
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("Total count: %d\n", counter.Load())
	// Output:
	// Value: 1
	// Value: 2
	// Value: 3
	// Total count: 3
}
