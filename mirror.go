// Package conch provides stream processing utilities for Go applications.
// This file implements stream mirroring functionality for replicating data flows.
package conch

import (
	"context"
	"sync"
)

// replicaPanicMsg is the error message displayed when invalid replica count is provided.
const replicaPanicMsg = "invalid mirror replicaCount"

// tee splits a single input stream into two identical output streams.
// It reads from inStream and sends each value to both output channels.
// The function uses a select statement to ensure both outputs receive the same data
// while respecting context cancellation.
func tee[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan T,
) (_, _ <-chan T) {
	outStream1 := make(chan T)
	outStream2 := make(chan T)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(outStream1)
		defer close(outStream2)

		for val := range inStream {
			out1, out2 := outStream1, outStream2

			for i := 0; i < 2; i++ {
				select {
				case <-ctx.Done():
					return
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()

	return outStream1, outStream2
}

// recBuildTeeTree recursively builds a binary tree structure for stream mirroring.
// It optimizes the mirroring process by creating a balanced tree that minimizes
// latency variance between output streams. This function is used internally
// by MirrorLowLatency to ensure homogeneous latency distribution.
func recBuildTeeTree[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	input <-chan T,
	outputs []<-chan T,
) {
	lo := len(outputs)

	if lo == 2 {
		outputs[0], outputs[1] = tee(ctx, wg, input)
		return
	}

	if lo == 3 {
		var ot <-chan T
		outputs[0], ot = tee(ctx, wg, input)
		outputs[1], outputs[2] = tee(ctx, wg, ot)

		return
	}

	o1, o2 := tee(ctx, wg, input)
	n1 := lo / 2
	recBuildTeeTree(ctx, wg, o1, outputs[:n1])
	recBuildTeeTree(ctx, wg, o2, outputs[n1:])
}

// MirrorHighThroughput replicate input stream to multiple output streams
// with maximum throughput.
//
//	Because its design requires fewer copy steps in the linear chain of flows
//	this version is faster. But it induces a greater latency disparity between
//	the input stream and the outputs. Latency is proportional to the rank of the
//	output stream.
//
//	See MirrorLowLatency if you need the lowest latency between input stream and
//	outputs.
func MirrorHighThroughput[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	replicaCount int,
	inStream <-chan T,
) []<-chan T {
	if replicaCount <= 0 {
		panic(replicaPanicMsg)
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	outStreams := make([]<-chan T, replicaCount)

	last := inStream

	for i := 0; i < replicaCount-1; i++ {
		outStreams[i], last = tee(ctx, wg, last)
	}

	outStreams[len(outStreams)-1] = last

	return outStreams
}

// MirrorsHighThroughputC creates a chainable function for high-throughput stream mirroring.
// It wraps MirrorHighThroughput to work with the ChainFunc interface, allowing it to be
// composed with other stream operations in a processing pipeline.
func MirrorsHighThroughputC[T any](
	replicas int,
	chains ChainsFunc[T],
) ChainFunc[T] {
	return func(ctx context.Context, wg *sync.WaitGroup, inStream <-chan T) {
		chains(ctx, wg, MirrorHighThroughput(ctx, wg, replicas, inStream)...)
	}
}

// MirrorLowLatency replicate input stream to multiple output streams with
// minimum latency.
//
//	Because its design uses a binary tree to separate the input stream from the
//	output streams, this version offers a homogeneous latency. But it induces more
//	copying steps and therefore a lower throughput.
//
//	See MirrorHighThroughput if you need the highest rate between input stream
//	and outputs.
func MirrorLowLatency[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	replicaCount int,
	inStream <-chan T,
) []<-chan T {
	if replicaCount <= 0 {
		panic(replicaPanicMsg)
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	inStreamBrk := inStream

	outStreams := make([]<-chan T, replicaCount)
	recBuildTeeTree(ctx, wg, inStreamBrk, outStreams)

	return outStreams
}

// MirrorsLowLatencyC creates a chainable function for low-latency stream mirroring.
// It wraps MirrorLowLatency to work with the ChainFunc interface, allowing it to be
// composed with other stream operations in a processing pipeline.
func MirrorsLowLatencyC[T any](
	replicas int,
	chains ChainsFunc[T],
) ChainFunc[T] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	) {
		chains(ctx, wg, MirrorLowLatency(ctx, wg, replicas, inStream)...)
	}
}
