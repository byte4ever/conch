// Package main demonstrates enriching items with metadata using decorators.
// This example shows how to use a decorator to add metadata to each data item.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/byte4ever/conch"
)

// Item represents a data item with metadata
// Metadata will be added by the decorator
type Item struct {
	ID      int
	Content string
	Meta    map[string]string
}

// MetadataDecorator adds metadata to each item
func MetadataDecorator(ctx context.Context, item *Item) {
	const errCtx = "metadata decoration"
	
	item.Meta = map[string]string{
		"timestamp": time.Now().Format(time.RFC3339),
		"source":    "decorator-example",
	}
	fmt.Printf("Decorated item %d with metadata\n", item.ID)
}

func main() {
	const errCtx = "main execution"
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	
	// Create a channel to inject data items
	injectFunc, inputStream := conch.Injector[Item](ctx)
	
	// Convert items to pointers for the decorator
	pointerStream := make(chan *Item)
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
			case item, more := <-inputStream:
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
	
	// Decorate each item with additional metadata
	decoratedStream := conch.Decorator(ctx, &wg, MetadataDecorator, pointerStream)
	
	// Create a pool for memory management
	pool := conch.NewPool[Item](10)
	
	// Consume the decorated items
	conch.Sink(ctx, &wg, pool, func(ctx context.Context, item *Item) {
		fmt.Printf("Processed item %d: %+v\n", item.ID, *item)
	}, decoratedStream)
	
	// Start feeding data into the pipeline
	fmt.Println("Starting decorator pipeline...")
	for i := 0; i < 3; i++ {
		item := Item{
			ID:      i,
			Content: fmt.Sprintf("Data content %d", i),
		}
		if err := injectFunc(ctx, item); err != nil {
			fmt.Printf("%s: failed to inject item %d: %v\n", errCtx, i, err)
			break
		}
	}
	
	// Wait for all processing to complete
	wg.Wait()
	fmt.Println("Decorator pipeline completed")
}
