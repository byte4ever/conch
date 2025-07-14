// Package main demonstrates the batch → process → unbatch pattern using conch.
// This example shows how to group individual items into batches, process them,
// and then unbatch them back to individual items.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

// BatchProcessor simulates processing a batch of integers
func BatchProcessor(ctx context.Context, batch []int) []int {
	const errCtx = "batch processing"
	
	// Simulate some processing time
	select {
	case <-ctx.Done():
		// Return empty batch if context is cancelled
		return []int{}
	case <-time.After(100 * time.Millisecond):
		// Processing completed
	}
	
	// Transform each item in the batch (multiply by 2)
	processed := make([]int, len(batch))
	for i, item := range batch {
		processed[i] = item * 2
	}
	
	fmt.Printf("Processed batch of %d items\n", len(batch))
	return processed
}

func main() {
	const errCtx = "main execution"
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	
	// Create injector to feed data
	injectFunc, inputStream := conch.Injector[int](ctx)
	
	// Step 1: Batch the input stream (batch size of 3)
	batchedStream := conch.Batch(ctx, &wg, 3, inputStream)
	
	// Step 2: Process batches
	processedStream := make(chan []int)
	wg.Add(1)
	go func() {
		defer func() {
			close(processedStream)
			wg.Done()
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			case batch, more := <-batchedStream:
				if !more {
					return
				}
				
				processed := BatchProcessor(ctx, batch)
				
				select {
				case <-ctx.Done():
					return
				case processedStream <- processed:
				}
			}
		}
	}()
	
	// Step 3: Unbatch the processed results
	unbatchedStream := conch.Unbatch(ctx, &wg, processedStream)
	
	// Step 4: Convert to pointer stream and consume the final results
	pointerStream := make(chan *int)
	wg.Add(1)
	go func() {
		defer func() {
			close(pointerStream)
			wg.Done()
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			case item, more := <-unbatchedStream:
				if !more {
					return
				}
				
				select {
				case <-ctx.Done():
					return
				case pointerStream <- &item:
				}
			}
		}
	}()
	
	pool := conch.NewPool[int](10)
	conch.Sink(ctx, &wg, pool, func(ctx context.Context, item *int) {
		fmt.Printf("Final result: %d\n", *item)
	}, pointerStream)
	
	// Feed data into the pipeline
	fmt.Println("Starting batch processing pipeline...")
	for i := 1; i <= 10; i++ {
		if err := injectFunc(ctx, i); err != nil {
			log.Printf("%s: failed to inject item %d: %v", errCtx, i, err)
			break
		}
	}
	
	// Wait for all processing to complete
	wg.Wait()
	fmt.Println("Batch processing pipeline completed")
}
