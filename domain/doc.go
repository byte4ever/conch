// Package domain provides core interface definitions and domain types for the
// conch stream processing library. This package establishes the fundamental
// contracts and abstractions that enable type-safe stream processing.
//
// Goals:
// - Define clear interfaces for processable elements and processors.
// - Establish type-safe contracts for cache operations.
// - Provide fundamental types for stream processing pipelines.
//
// Design Philosophy:
// The domain package follows the principle of separating interface definitions
// from implementation details. It defines what operations are possible without
// dictating how they should be implemented. This allows for multiple
// implementations while maintaining type safety and clear contracts.
//
// Key interfaces include:
// - Processable: Elements that can be processed in a stream pipeline.
// - Cache: Generic caching interface for stream processing optimization.
// - Hashable: Elements that can be hashed for caching and indexing.
//
// Usage Patterns:
// - Implement Processable for stream elements that need processing.
// - Use Cache interface for implementing caching strategies.
// - Implement Hashable for elements that need to be cached or indexed.
//
// Example:
//
// package main
//
// import (
//     "context"
//     "fmt"
//     "github.com/byte4ever/conch/domain"
// )
//
// type MyProcessable struct {
//     param  string
//     result string
//     err    error
// }
//
// func (p *MyProcessable) GetParam() string { return p.param }
// func (p *MyProcessable) SetValue(v string) { p.result = v; p.err = nil }
// func (p *MyProcessable) GetValue() string { return p.result }
// func (p *MyProcessable) SetError(e error) { p.err = e; p.result = "" }
// func (p *MyProcessable) GetError() error { return p.err }
//
// func exampleProcessor(ctx context.Context, elem *MyProcessable) {
//     // Process the element
//     result := "processed: " + elem.GetParam()
//     elem.SetValue(result)
// }
//
// func main() {
//     ctx := context.Background()
//     elem := &MyProcessable{param: "test"}
//     
//     exampleProcessor(ctx, elem)
//     
//     fmt.Println(elem.GetValue())
// }
package domain
