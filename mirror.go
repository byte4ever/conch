package conch

import (
	"context"
	"sync"
)

const replicaPanicMsg = "invalid mirror replicaCount"

func tee[T any](wg *sync.WaitGroup, inStream <-chan T) (_, _ <-chan T) {
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

func recBuildTeeTree[T any](wg *sync.WaitGroup, input <-chan T, outputs []<-chan T) {
	lo := len(outputs)

	if lo == 2 {
		outputs[0], outputs[1] = tee(wg, input)
		return
	}

	if lo == 3 {
		var ot <-chan T
		outputs[0], ot = tee(wg, input)
		outputs[1], outputs[2] = tee(wg, ot)

		return
	}

	o1, o2 := tee(wg, input)
	n1 := lo / 2
	recBuildTeeTree(wg, o1, outputs[:n1])
	recBuildTeeTree(wg, o2, outputs[n1:])
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
	if replicaCount == 0 {
		panic(replicaPanicMsg)
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	outStreams := make([]<-chan T, replicaCount)

	last := ContextBreaker(ctx, inStream)

	for i := 0; i < replicaCount-1; i++ {
		outStreams[i], last = tee(wg, last)
	}

	outStreams[len(outStreams)-1] = last

	return outStreams
}

func MirrorHighThroughputC[T any](
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
// Because its design uses a binary tree to separate the input stream from the
// output streams, this version offers a homogeneous latency. But it induces
// more copying steps and therefore a lower throughput.
//
//	See MirrorHighThroughput if you need the highest rate between input stream
//	and outputs.
func MirrorLowLatency[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	replicaCount int,
	inStream <-chan T,
) []<-chan T {
	if replicaCount == 0 {
		panic(replicaPanicMsg)
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	inStreamBrk := ContextBreaker(ctx, inStream)

	outStreams := make([]<-chan T, replicaCount)
	recBuildTeeTree(wg, inStreamBrk, outStreams)

	return outStreams
}

func MirrorLowLatencyC[T any](
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
