// Package valve provides valve control mechanisms for stream processing
// applications. It offers simple on/off control for stream flow management.
//
// Goals:
// - Provide simple valve control for stream flow management.
// - Enable dynamic control of stream processing pipelines.
// - Support context-aware valve operations.
//
// Design Philosophy:
// The valve package provides a simple yet effective way to control stream
// flow through boolean signaling. It follows the principle of simplicity
// and can be used to pause, resume, or conditionally process streams based
// on external conditions.
//
// Usage Patterns:
// - Use valves to control stream processing flow.
// - Open valves to allow stream processing.
// - Close valves to pause or stop stream processing.
// - Monitor valve state through channels for reactive control.
//
// Example:
//
// package main
//
// import (
//     "context"
//     "fmt"
//     "github.com/byte4ever/conch/valve"
// )
//
// func main() {
//     ctx := context.Background()
//     
//     // Create a new valve
//     v := valve.New()
//     
//     // Open the valve
//     v.Open()
//     
//     // Get the valve channel
//     valveChannel := v.GetChannel(ctx)
//     
//     // Check valve state
//     select {
//     case state := <-valveChannel:
//         if state {
//             fmt.Println("Valve is open")
//         } else {
//             fmt.Println("Valve is closed")
//         }
//     default:
//         fmt.Println("No valve state change")
//     }
// }
package valve
