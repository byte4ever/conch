# Conch

Conch is a powerful Golang library for implementing a variety of stream-based concurrency patterns. It focuses on stream processing, functional chaining, and context-aware concurrency, providing developers the flexibility to build complex data pipelines with ease.

## Key Concepts

- **Stream Processing**: Leverage streams to process data asynchronously with low latency.
- **Functional Chaining**: Build complex logic by chaining simple, reusable functions.
- **Context-Aware Concurrency**: Utilize Go's `context` to manage goroutines efficiently and responsively.

## Features

- **Batch/Unbatch**: Efficiently group and ungroup data items.
- **Mirror**: Duplicate streams into multiple output paths for extensive processing.
- **Decorator**: Apply transformations or enhancements to data items.
- **Injector**: Inject data into streams dynamically.
- **Cache**: Integrate seamless caching mechanisms to optimize repeated operations.
- **Multiplex**: Merge or split data streams effectively.
- **Circuit Breaker & Retrier**: Integrate resilience patterns to manage failure and retry strategies.

## Installation

To include Conch in your project, run:

```bash
go get github.com/byte4ever/conch
```

## Quick-start Code Snippet

Here's a simple example to create a data pipeline with Conch:

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "github.com/byte4ever/conch"
)

func main() {
    ctx := context.Background()
    wg := sync.WaitGroup{}

    injectFunc, inStream := conch.Injector[int](ctx)

    outStream := conch.Sink(ctx, &wg, nil, func(ctx context.Context, item *int) {
        fmt.Println(*item)
    }, inStream)

    for i := 0; i < 10; i++ {
        _ = injectFunc(ctx, i)
    }

    wg.Wait()
}
```

## Detailed Usage Examples

- **Batching and Unbatching**: Group or split data based on custom criteria.
- **Mirroring for Parallel Processing**: Send data to multiple processing streams in parallel.
- **Resiliency Patterns with Circuit Breaker**: Protect your system from failure excess.

## API Surface Overview

The complete API documentation can be hosted on [GoDoc](https://pkg.go.dev/github.com/byte4ever/conch).

## Contribution, Coding Style, Linting, Test Instructions

We welcome contributions! Please adhere to the following:

- **Code Style**: Follow idiomatic Go practices and structure.
- **Linting**: Use `golangci-lint` with our configuration located at `.golangci.yaml`.
- **Testing**: Run tests using `go test ./...`. Ensure comprehensive test coverage for new features or patches.

## License

Conch is licensed under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details.

## CI Status

[![Build Status](https://travis-ci.org/byte4ever/conch.svg?branch=main)](https://travis-ci.org/byte4ever/conch)
