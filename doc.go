// Package conch provides a set of tools and utilities to implement stream-based
// data processing patterns in Go. The library is designed to facilitate the
// development of high-performance and scalable stream pipelines with focus on
// concurrency and functional programming paradigms.
//
// Goals:
// - Simplify stream processing in Go.
// - Enable flexible composition of processing functions.
// - Promote context-aware concurrent execution.
//
// Design Philosophy:
// Conch divides responsibilities into small, reusable components like
// Injectors, Processors, and Sinks. These components can be chained together
// to form complex data processing pipelines. The emphasis is on composability,
// non-blocking execution, and making maximal use of Go's concurrency features.
//
// Usage Patterns:
// - Use Injectors to introduce data into a pipeline.
// - Chain processing functions to manipulate each data item.
// - Employ Sinks to consume the final output from the pipeline.
//
// See the run example for a quick demonstration of a simple pipeline.
//
// Example:
//
// package main
//
// import (
//     "context"
//     "fmt"
//     "sync"
//     "github.com/byte4ever/conch"
// )
//
// func main() {
//     ctx := context.Background()
//     wg := sync.WaitGroup{}
//
//     injectFunc, inStream := conch.Injector[int](ctx)
//
//     outStream := conch.Sink(ctx, &wg, nil, func(ctx context.Context, item *int) {
//         fmt.Println(*item)
//     }, inStream)
//
//     for i := 0; i < 10; i++ {
//         _ = injectFunc(ctx, i)
//     }
//
//     wg.Wait()
// }
package conch
