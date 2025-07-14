// Package retrier provides retry mechanisms for stream processing applications.
// It offers configurable retry policies with exponential backoff and jitter
// to handle transient failures gracefully.
//
// Goals:
// - Provide configurable retry policies for failed operations.
// - Support exponential backoff with jitter for retry intervals.
// - Enable integration with stream processors for resilient processing.
//
// Design Philosophy:
// The retrier package focuses on graceful handling of transient failures
// through configurable retry strategies. It employs exponential backoff with
// jitter to prevent thundering herd problems and allows fine-tuning of retry
// behavior based on application requirements.
//
// Usage Patterns:
// - Wrap processing functions with retry logic for fault tolerance.
// - Configure retry policies with maximum attempts and duration limits.
// - Use RetryableError to indicate errors that should trigger retry attempts.
//
// Example:
//
// package main
//
// import (
//     "context"
//     "fmt"
//     "github.com/byte4ever/conch/retrier"
// )
//
// func main() {
//     ctx := context.Background()
//     
//     // Create a retrier with maximum 3 attempts
//     r, err := retrier.New(
//         retrier.WithMaxRetry(3),
//         retrier.WithBackoffFactor(2.0),
//     )
//     if err != nil {
//         fmt.Println("Error creating retrier:", err)
//         return
//     }
//
//     // Wrap a processor function with retry logic
//     processor := retrier.WrapProcessorFunc(r, func(ctx context.Context, elem *int) {
//         fmt.Println("Processing:", *elem)
//     })
//
//     elem := 1
//     processor(ctx, &elem)
// }
package retrier
