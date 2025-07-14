# Conch

Conch is a powerful Golang library for implementing a variety of stream-based concurrency patterns. It focuses on stream processing, functional chaining, and context-aware concurrency, providing developers the flexibility to build complex data pipelines with ease.

## Why Choose Conch?

### 🚀 **Production-Ready Performance**
- **Zero-allocation streaming**: Efficient memory management with built-in object pooling
- **Backpressure handling**: Automatic flow control prevents memory exhaustion
- **Optimized mirroring**: Choose between high-throughput or low-latency stream duplication
- **Concurrent processing**: Built-in goroutine management with proper cleanup

### 🛠️ **Developer Experience**
- **Type-safe generics**: Full Go 1.18+ generic support eliminates runtime casting
- **Context-aware**: Native `context.Context` integration for cancellation and timeouts
- **Composable patterns**: Mix and match components to build complex pipelines
- **Minimal boilerplate**: Clean, readable APIs that reduce development time

### 🔧 **Real-World Problem Solving**
- **Batch processing**: Handle high-volume data efficiently with configurable batching
- **Stream mirroring**: Send data to multiple services (logging, storage, analytics) simultaneously
- **Data enrichment**: Add metadata, transform, or validate data in-flight
- **External integration**: Seamlessly inject data from APIs, databases, or message queues
- **Resilience patterns**: Built-in circuit breakers and retry mechanisms

### 💡 **When to Use Conch**
- Building **data pipelines** for ETL/ELT processes
- Implementing **event-driven architectures** with multiple consumers
- Creating **microservices** that need to process streams of data
- Building **real-time analytics** systems
- Handling **high-throughput APIs** with concurrent processing
- Implementing **fan-out/fan-in** patterns for distributed processing

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

## Getting Started in 5 Minutes

### Step 1: Basic Pipeline
Create a simple data processing pipeline:

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
    
    // Create data injector
    inject, stream := conch.Injector[string](ctx)
    
    // Add data processor
    processed := conch.Decorator(ctx, &wg, func(ctx context.Context, item *string) {
        *item = fmt.Sprintf("processed: %s", *item)
    }, stream)
    
    // Add output sink
    conch.Sink(ctx, &wg, nil, func(ctx context.Context, item *string) {
        fmt.Println(*item)
    }, processed)
    
    // Feed data
    inject(ctx, "hello")
    inject(ctx, "world")
    
    wg.Wait()
}
```

### Step 2: Add Parallel Processing
Scale up with stream mirroring:

```go
// Mirror to multiple processors
streams := conch.MirrorHighThroughput(ctx, &wg, 3, inputStream)

// Process in parallel
conch.Sink(ctx, &wg, pool, logProcessor, streams[0])
conch.Sink(ctx, &wg, pool, saveProcessor, streams[1])
conch.Sink(ctx, &wg, pool, analyzeProcessor, streams[2])
```

### Step 3: Add Batching for High Volume
Optimize for throughput:

```go
// Batch items for efficient processing
batched := conch.Batch(ctx, &wg, 100, inputStream)
processed := processBatches(ctx, &wg, batched)
unbatched := conch.Unbatch(ctx, &wg, processed)
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

## Runnable Examples

Explore practical implementations in the `examples/` directory:

### 1. Batch Processing Pipeline
```bash
go run ./examples/batching/main.go
```
Demonstrates the **batch → process → unbatch** pattern:
- Groups individual items into batches of 3
- Processes each batch (multiplies by 2)
- Unbatches results back to individual items
- Perfect for high-volume data processing

### 2. Stream Mirroring
```bash
go run ./examples/mirroring/main.go
```
Shows how to **mirror streams to multiple services**:
- Duplicates a single stream to logging and storage services
- Processes both streams in parallel
- Ideal for event-driven architectures

### 3. Data Enrichment
```bash
go run ./examples/decorator/main.go
```
Illustrates **enriching items with metadata**:
- Adds timestamps and source information to data items
- Demonstrates in-flight data transformation
- Useful for data validation and enrichment pipelines

### 4. External Data Injection
```bash
go run ./examples/injector/main.go
```
Shows **external producer injecting into pipeline**:
- Simulates external API feeding data into the system
- Demonstrates real-time data ingestion patterns
- Perfect for microservices integration

## Performance Comparison

```go
// Traditional approach - blocking and error-prone
for _, item := range items {
    result1 := service1.Process(item)  // Sequential
    result2 := service2.Process(item)  // Sequential
    // Error handling mixed with business logic
}

// Conch approach - concurrent and clean
injectFunc, stream := conch.Injector[Item](ctx)
mirrored := conch.MirrorHighThroughput(ctx, &wg, 2, stream)

conch.Sink(ctx, &wg, pool, service1.Process, mirrored[0])  // Parallel
conch.Sink(ctx, &wg, pool, service2.Process, mirrored[1])  // Parallel
```

## Real-World Use Cases

### E-commerce Order Processing
```go
// Process orders, send to inventory, payment, and notification services
orders := conch.Injector[Order](ctx)
streams := conch.MirrorHighThroughput(ctx, &wg, 3, orders)

conch.Sink(ctx, &wg, pool, inventoryService.Reserve, streams[0])
conch.Sink(ctx, &wg, pool, paymentService.Charge, streams[1])
conch.Sink(ctx, &wg, pool, notificationService.Send, streams[2])
```

### Log Processing Pipeline
```go
// Batch logs, process them, and store in multiple formats
logs := conch.Injector[LogEntry](ctx)
batched := conch.Batch(ctx, &wg, 100, logs)
processed := conch.Decorator(ctx, &wg, enrichWithMetadata, batched)
unbatched := conch.Unbatch(ctx, &wg, processed)

conch.Sink(ctx, &wg, pool, elasticsearchWriter.Write, unbatched)
```

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
