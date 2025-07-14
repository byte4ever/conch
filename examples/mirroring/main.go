// Package main demonstrates stream mirroring to multiple services using conch.
// This example shows how to mirror a single stream to both a logging service
// and a storage service for parallel processing.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

// LoggingService simulates a logging service that processes data
type LoggingService struct {
	name string
}

func (ls *LoggingService) Process(ctx context.Context, item *int) {
	const errCtx = "logging service processing"
	
	// Simulate logging processing time
	select {
	case <-ctx.Done():
		return
	case <-time.After(50 * time.Millisecond):
		// Processing completed
	}
	
	fmt.Printf("[%s] Logged item: %d\n", ls.name, *item)
}

// StorageService simulates a storage service that persists data
type StorageService struct {
	name string
	data []int
	mu   sync.Mutex
}

func (ss *StorageService) Store(ctx context.Context, item *int) {
	const errCtx = "storage service processing"
	
	// Simulate storage processing time
	select {
	case <-ctx.Done():
		return
	case <-time.After(100 * time.Millisecond):
		// Processing completed
	}
	
	ss.mu.Lock()
	ss.data = append(ss.data, *item)
	ss.mu.Unlock()
	
	fmt.Printf("[%s] Stored item: %d (total: %d items)\n", ss.name, *item, len(ss.data))
}

func (ss *StorageService) GetStoredData() []int {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	result := make([]int, len(ss.data))
	copy(result, ss.data)
	return result
}

func main() {
	const errCtx = "main execution"
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	
	// Create injector to feed data
	injectFunc, inputStream := conch.Injector[int](ctx)
	
	// Mirror the input stream to 2 output streams
	mirroredStreams := conch.MirrorHighThroughput(ctx, &wg, 2, inputStream)
	
	// Create services
	logger := &LoggingService{name: "MainLogger"}
	storage := &StorageService{name: "PrimaryStorage", data: make([]int, 0)}
	
	// Create pools for each service
	logPool := conch.NewPool[int](10)
	storagePool := conch.NewPool[int](10)
	
	// Convert first mirrored stream to pointer stream and connect to logging service
	logPointerStream := make(chan *int)
	wg.Add(1)
	go func() {
		defer func() {
			close(logPointerStream)
			wg.Done()
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			case item, more := <-mirroredStreams[0]:
				if !more {
					return
				}
				
				select {
				case <-ctx.Done():
					return
				case logPointerStream <- &item:
				}
			}
		}
	}()
	
	conch.Sink(ctx, &wg, logPool, func(ctx context.Context, item *int) {
		logger.Process(ctx, item)
	}, logPointerStream)
	
	// Convert second mirrored stream to pointer stream and connect to storage service
	storagePointerStream := make(chan *int)
	wg.Add(1)
	go func() {
		defer func() {
			close(storagePointerStream)
			wg.Done()
		}()
		
		for {
			select {
			case <-ctx.Done():
				return
			case item, more := <-mirroredStreams[1]:
				if !more {
					return
				}
				
				select {
				case <-ctx.Done():
					return
				case storagePointerStream <- &item:
				}
			}
		}
	}()
	
	conch.Sink(ctx, &wg, storagePool, func(ctx context.Context, item *int) {
		storage.Store(ctx, item)
	}, storagePointerStream)
	
	// Start feeding data
	fmt.Println("Starting mirroring pipeline...")
	
	// Feed data into the pipeline
	for i := 1; i <= 5; i++ {
		if err := injectFunc(ctx, i*10); err != nil {
			log.Printf("%s: failed to inject item %d: %v", errCtx, i*10, err)
			break
		}
		
		// Add some delay between injections to observe parallel processing
		time.Sleep(50 * time.Millisecond)
	}
	
	// Wait for all processing to complete
	wg.Wait()
	
	// Display final results
	fmt.Println("\nMirroring pipeline completed")
	fmt.Printf("Storage service collected data: %v\n", storage.GetStoredData())
	fmt.Println("Both logging and storage services processed all items in parallel")
}
