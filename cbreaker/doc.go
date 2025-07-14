// Package cbreaker provides circuit breaker implementations to enhance the
// resilience of stream processing applications. Circuit breakers prevent
// repetitive failures and manage state transitions to ensure system stability.
//
// Goals:
// - Offer a circuit breaker mechanism for controlling failures.
// - Manage state transitions between Closed, Open, and HalfOpen states.
// - Support resilient integration with stream processors.
//
// Design Philosophy:
// The cbreaker package focuses on managing transient faults and avoiding
// recurrent failures in distributed systems. It allows systems to fail fast,
// prevent cascading failures, and gracefully transition between states based
// on success and failure counts.
//
// Usage Patterns:
// - Wrap processing functions with a circuit breaker to control execution flow.
// - Use state transitions to manage and recover from failures.
// - Configure thresholds and timeouts to customize circuit breaker behavior.
//
// Example:
//
// package main
//
// import (
//     "context"
//     "fmt"
//     "github.com/byte4ever/conch/cbreaker"
// )
//
// func main() {
//     engine, err := cbreaker.NewEngine()
//     if err != nil {
//         fmt.Println("Error creating circuit breaker:", err)
//         return
//     }
//
//     // Simulated processing function with circuit breaker
//     process := cbreaker.WrapProcessorFunc(engine, func(ctx context.Context, elem *int) {
//         fmt.Println("Processing:", *elem)
//     })
//
//     elem := 1
//     ctx := context.Background()
//     
//     process(ctx, &elem)
//     
//     if engine.IsOpen() {
//         fmt.Println("Circuit breaker open")
//     } else {
//         fmt.Println("Circuit breaker closed")
//     }
// }
package cbreaker
