package conch_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/byte4ever/conch"
)

// Example demonstrates a minimal pipeline using conch.
func Example() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	// Create an injector to feed data into the pipeline
	injectFunc, inStream := conch.Injector[int](ctx)

	// Create a sink to consume the output
	conch.Sink(ctx, &wg, nil, func(ctx context.Context, item *int) {
		fmt.Println(*item)
	}, inStream)

	// Inject some data into the pipeline
	for i := 0; i < 3; i++ {
		_ = injectFunc(ctx, i)
	}

	wg.Wait()
	// Output:
	// 0
	// 1
	// 2
}
