// Package main demonstrates external producer injecting into pipeline using conch.
// This example shows how to use an injector to feed data from an external source
// into a processing pipeline.
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

// ExternalDataSource simulates an external data source
type ExternalDataSource struct {
	name    string
	counter int
}

// GenerateData simulates generating data from an external source
func (eds *ExternalDataSource) GenerateData(ctx context.Context) (string, error) {
	const errCtx = "external data generation"
	
	// Simulate some processing time
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("%s: context cancelled", errCtx)
	case <-time.After(200 * time.Millisecond):
		// Processing completed
	}
	
	eds.counter++
	data := fmt.Sprintf("Data-%s-%d", eds.name, eds.counter)
	fmt.Printf("Generated data: %s\n", data)
	return data, nil
}

// DataProcessor processes incoming data
func DataProcessor(ctx context.Context, data *string) {
	const errCtx = "data processing"
	
	// Simulate processing time
	select {
	case <-ctx.Done():
		return
	case <-time.After(100 * time.Millisecond):
		// Processing completed
	}
	
	processed := fmt.Sprintf("PROCESSED[%s]", *data)
	fmt.Printf("Processed: %s\n", processed)
}

// ExternalProducer simulates an external producer that feeds data into the pipeline
func ExternalProducer(ctx context.Context, injector func(context.Context, string) error, source *ExternalDataSource) {
	const errCtx = "external producer"
	
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			fmt.Println("External producer stopped")
			return
		case <-ticker.C:
			data, err := source.GenerateData(ctx)
			if err != nil {
				log.Printf("%s: failed to generate data: %v", errCtx, err)
				continue
			}
			
			if err := injector(ctx, data); err != nil {
				log.Printf("%s: failed to inject data: %v", errCtx, err)
				continue
			}
		}
	}
}

func main() {
	const errCtx = "main execution"
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	
	// Create injector for the pipeline
	injectFunc, inputStream := conch.Injector[string](ctx)
	
	// Convert to pointer stream for processing
	pointerStream := make(chan *string)
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
			case data, more := <-inputStream:
				if !more {
					return
				}
				
				select {
				case <-ctx.Done():
					return
				case pointerStream <- &data:
				}
			}
		}
	}()
	
	// Create a pool for memory management
	pool := conch.NewPool[string](10)
	
	// Set up data processor as sink
	conch.Sink(ctx, &wg, pool, DataProcessor, pointerStream)
	
	// Create external data source
	dataSource := &ExternalDataSource{name: "ExternalAPI"}
	
	// Start external producer in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ExternalProducer(ctx, injectFunc, dataSource)
	}()
	
	fmt.Println("Starting external producer pipeline...")
	fmt.Println("External producer will feed data into the pipeline for 5 seconds...")
	
	// Wait for all processing to complete
	wg.Wait()
	fmt.Println("External producer pipeline completed")
}
