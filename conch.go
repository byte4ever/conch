package conch

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/byte4ever/conch/internal/ratelimit"
)

type Generator[T any] func(context.Context) (output <-chan T, err error)

func DrainDo[T any](
	ctx context.Context,
	inStream <-chan T,
	do func(context.Context, T) error,
) error {
	for v := range inStream {
		if do != nil {
			err := do(ctx, v)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// ContextBreaker creates an output stream that copy input stream and
// that close when the context is done or input stream is closed.
//
//	This is useful to cut properly stream flows, especially when down stream
//	enter some operators that are no longer sensitive to context termination.
func ContextBreaker[T any](
	ctx context.Context,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for t := range inStream {
			select {
			case <-ctx.Done():
				return

			case outStream <- t:
			}
		}
	}()

	return outStream
}

// WorkerPool launch the given generators concurrently and multiplexes the
// outputs in a single output stream.
func WorkerPool[T any](
	ctx context.Context,
	egs ...Generator[T],
) (outputStream <-chan T, err error) {
	le := len(egs)
	if le == 0 {
		return nil, nil //nolint:nilnil
	}

	inStreams := make([]<-chan T, le)

	for i, eg := range egs {
		inStreams[i], err = eg(ctx)
		if err != nil {
			return nil, err
		}
	}

	return FanIn(ctx, inStreams...), nil
}

// FanIn multiplexes input streams into a single output stream with balanced
// priority.
//
//	Output stream is closed when the context is done or all input streams are
//	closed.
func FanIn[T any](
	ctx context.Context,
	inStreams ...<-chan T,
) <-chan T {
	var wg sync.WaitGroup

	wg.Add(len(inStreams))

	multiplexedOutStream := make(chan T, len(inStreams))

	for _, c := range inStreams {
		go multiplex(ctx, &wg, c, multiplexedOutStream)
	}

	go func() {
		wg.Wait()
		close(multiplexedOutStream)
	}()

	return multiplexedOutStream
}

// fairMerge merges two input streams into a single output stream with
// balanced priority
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func fairMerge[T any](
	inStream1, inStream2 <-chan T,
) chan T {
	multiplexedOutStream := make(chan T)

	go func() {
		defer close(multiplexedOutStream)

		i1, i2 := inStream1, inStream2

		for {
			if i1 == nil && i2 == nil {
				return
			}
			select {
			case t, more := <-i1:
				if !more {
					i1 = nil
					continue
				}
				multiplexedOutStream <- t
			case t, more := <-inStream2:
				if !more {
					i2 = nil
					continue
				}
				multiplexedOutStream <- t
			}
		}
	}()

	return multiplexedOutStream
}

// unfairMerge merges two input streams into a single output stream with
// unbalanced priority.
//
//	Output stream is closed when both of them are closed.
//
//	Please note that the more back pressure is present on the output stream the
//	more priority effect apply.
func unfairMerge[T any](
	lowPrioInStream, highPrioInStream <-chan T,
) chan T {
	multiplexedOutStream := make(chan T)

	go func() {
		defer close(multiplexedOutStream)

		low, high := lowPrioInStream, highPrioInStream

		for {
			switch {
			case low == nil && high == nil:
				// both of them are closed so close the output stream
				return

			case low == nil:
				// low prio stream is closed, so we use high prio stream only
				t, more := <-high
				if !more {
					high = nil
					continue
				}
				multiplexedOutStream <- t

			case high == nil:
				// high prio stream is closed, so we use low prio stream only
				t, more := <-low
				if !more {
					low = nil
					continue
				}
				multiplexedOutStream <- t
			default:
				select {
				// we try higher prio stream first
				case t, more := <-high:
					if !more {
						high = nil
						continue
					}
					multiplexedOutStream <- t
				default:
					// otherwise we try both of them
					select {
					case t, more := <-high:
						if !more {
							high = nil
							continue
						}
						multiplexedOutStream <- t
					case t, more := <-low:
						if !more {
							low = nil
							continue
						}
						multiplexedOutStream <- t
					}
				}
			}
		}
	}()

	return multiplexedOutStream
}

func multiplex[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	inStream <-chan T,
	outStream chan T,
) {
	defer wg.Done()

	for t := range inStream {
		select {
		case <-ctx.Done():
			return
		case outStream <- t:
			// emit
		}
	}
}

func tee[T any](inStream <-chan T) (_, _ <-chan T) {
	outStream1 := make(chan T)
	outStream2 := make(chan T)

	go func() {
		defer close(outStream1)
		defer close(outStream2)

		for val := range inStream {
			var out1, out2 = outStream1, outStream2

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

func recBuildTeeTree[T any](input <-chan T, outputs []<-chan T) {
	lo := len(outputs)

	if lo == 2 {
		outputs[0], outputs[1] = tee(input)
		return
	}

	if lo == 3 {
		var ot <-chan T
		outputs[0], ot = tee(input)
		outputs[1], outputs[2] = tee(ot)
		return
	}

	o1, o2 := tee(input)
	n1 := lo / 2
	recBuildTeeTree(o1, outputs[:n1])
	recBuildTeeTree(o2, outputs[n1:])
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
	inStream <-chan T,
	replicaCount int,
) []<-chan T {
	if replicaCount == 0 {
		panic("invalid split replicaCount")
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	outStreams := make([]<-chan T, replicaCount)

	last := ContextBreaker(ctx, inStream)

	for i := 0; i < replicaCount-1; i++ {
		outStreams[i], last = tee(last)
	}

	outStreams[len(outStreams)-1] = last

	return outStreams
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
	inStream <-chan T,
	replicaCount int,
) []<-chan T {
	if replicaCount == 0 {
		panic("invalid split replicaCount")
	}

	if replicaCount == 1 {
		return []<-chan T{inStream}
	}

	inStreamBrk := ContextBreaker(ctx, inStream)

	outStreams := make([]<-chan T, replicaCount)
	recBuildTeeTree(inStreamBrk, outStreams)

	return outStreams
}

// FairFanIn merges multiple input streams into a single output stream with
// unbalanced priority.
//
//	Priority is defined by the input stream rank, from lowest to highest.
//	Function will panic if no input stream are provided.
//
//	The more back pressure is present on output stream (i.e. slow consumption)
//	the more unfair priority effect will occur.
//
//	If two input stream are provided, they will be merged into one output stream
//	with exact same priority.
//
//	If more are provided, stream in third position will have a twice the
//	throughput of the stream in second position.
//
//	throughput(inputSteam#n) = 2 * throughput(inputSteam#(n-1))
//
//	If less back pressure is present, the more priority is balanced.
//
//	Output stream is closed when all input stream are closed or context is done.
func FairFanIn[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	count := len(inStream)

	if count == 0 {
		panic("invalid priority count")
	}

	if count == 1 {
		input := make(chan T)
		return ContextBreaker(ctx, input)
	}

	// decouple from input stream because unfairMerge doesn't use context for
	// clean up. We need to ensure the funnel will be destroyed by channel
	// closing cascading effect only.
	bridgeInStreams := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		bridgeInStreams[i] = ContextBreaker(ctx, inStream[i])
	}

	// creates fair merge funnel
	outStream := fairMerge(bridgeInStreams[0], bridgeInStreams[1])
	for _, inStream := range bridgeInStreams[2:] {
		outStream = fairMerge(inStream, outStream)
	}

	return outStream
}

// UnfairFanIn merges multiple input streams into a single output stream with
// unfair priority balance.
//
// Priority is defined by the input stream rank, from lowest to highest.
// Function will panic if no input stream are provided.
//
// The more back pressure is present on output stream (i.e. slow consumption)
// the more priority effect will occur.
//
// * inputSteam#n is got less priority than inputSteam#(n+1)
//
// If no back pressure is present, the more priority looks  balanced.
//
// Output stream is closed when all input stream are closed or context is done.
func UnfairFanIn[T any](
	ctx context.Context,
	inStream ...<-chan T,
) <-chan T {
	count := len(inStream)

	if count == 0 {
		panic("invalid priority count")
	}

	if count == 1 {
		input := make(chan T)
		return ContextBreaker(ctx, input)
	}

	// decouple from input stream because unfairMerge doesn't use context for
	// clean up. We need to ensure the funnel will be destroyed by channel
	// closing cascading effect only.
	bridgeInStreams := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		bridgeInStreams[i] = ContextBreaker(
			ctx,
			inStream[i],
		)
	}

	// creates unfair merge funnel
	outStream := unfairMerge(bridgeInStreams[0], bridgeInStreams[1])
	for _, inStream := range bridgeInStreams[2:] {
		outStream = unfairMerge(outStream, inStream)
	}

	return outStream
}

type Comparable interface {
	LessThan(other Comparable) bool
}

type queue []Comparable

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].LessThan(q[j])
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *queue) Push(x any) {
	item := x.(Comparable)
	*q = append(*q, item)
}

func (q *queue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*q = old[0 : n-1]

	return item
}

type prioritizeOption struct {
	backPressureDelay time.Duration
	flushWhenDown     bool
	bufferSize        int
}

var defaultPrioritizeOption = prioritizeOption{
	backPressureDelay: 0,
	flushWhenDown:     false,
	bufferSize:        1000,
}

type PrioritizeOption interface {
	apply(*prioritizeOption)
}

type backPressureDelayOption time.Duration

func (b backPressureDelayOption) apply(option *prioritizeOption) {
	option.backPressureDelay = time.Duration(b)
}

func WithBackPressureDelay(d time.Duration) PrioritizeOption {
	return backPressureDelayOption(d)
}

type flushWhenDown struct{}

func (b flushWhenDown) apply(option *prioritizeOption) {
	option.flushWhenDown = true
}

func WithFlushWhenDown() PrioritizeOption {
	return flushWhenDown{}
}

type bufferSize int

func (b bufferSize) apply(option *prioritizeOption) {
	option.bufferSize = int(b)
}

func WithBufferSize(d int) PrioritizeOption {
	return bufferSize(d)
}

// Prioritize bufferize input stream into a priority queue and copy element
// in output stream by priority order.
func Prioritize[T Comparable](
	ctx context.Context,
	inStream <-chan T,
	option ...PrioritizeOption,
) <-chan T {
	var (
		mtx sync.Mutex
		wg  sync.WaitGroup
	)

	opt := defaultPrioritizeOption
	for _, o := range option {
		o.apply(&opt)
	}

	availabilityStream := make(chan struct{}, opt.bufferSize)
	pq := make(queue, 0, opt.bufferSize)
	outStream := make(chan T)

	wg.Add(2)

	innerCtx, innerCancel := context.WithCancel(ctx)

	go func() {
		defer wg.Done()

		delay := time.NewTimer(opt.backPressureDelay)

		for {
			select {
			case <-delay.C:
				for {
					select {
					case <-availabilityStream:
						mtx.Lock()
						item, _ := heap.Pop(&pq).(T) // nolint:errcheck
						mtx.Unlock()

						select {
						case outStream <- item:
						case <-innerCtx.Done():
							mtx.Lock()
							heap.Push(&pq, item)
							mtx.Unlock()
							availabilityStream <- struct{}{}

							return
						}

					case <-innerCtx.Done():
						return
					}
				}

			case <-innerCtx.Done():
				if !delay.Stop() {
					<-delay.C
				}
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		defer innerCancel()

		for {
			select {
			case <-ctx.Done():
				return
			case v, more := <-inStream:
				if !more {
					return
				}
				mtx.Lock()
				heap.Push(&pq, v)
				mtx.Unlock()

				availabilityStream <- struct{}{}
			}
		}
	}()

	go func() {
		wg.Wait()
		defer close(outStream)
		close(availabilityStream)

		if opt.flushWhenDown {
			for range availabilityStream {
				item, _ := heap.Pop(&pq).(T) // nolint:errcheck
				outStream <- item
			}
		}
	}()

	return outStream
}

type valveCtrl int8

const (
	valveOpenOP valveCtrl = iota
	valveCloseOP
)

type valve struct {
	blockingCtx context.Context
	ctrl        chan valveCtrl
}

func ClosedValve[T any](
	ctx context.Context, inStream <-chan T,
) (func(), func(), <-chan T) {
	var op valveCtrl
	op = valveCloseOP

	outStream, open, _close := buildValve(ctx, inStream, op)
	return open, _close, outStream
}

func OpenedValve[T any](
	ctx context.Context, inStream <-chan T,
) (func(), func(), <-chan T) {
	var op valveCtrl
	op = valveOpenOP

	outStream, open, _close := buildValve(ctx, inStream, op)
	return open, _close, outStream
}

func buildValve[T any](
	ctx context.Context,
	inStream <-chan T,
	op valveCtrl,
) (
	<-chan T, func(), func(),
) {
	outStream := make(chan T)
	blockingCtx, blockingCancel := context.WithCancel(ctx)

	innerValve := &valve{
		blockingCtx: blockingCtx,
		ctrl:        make(chan valveCtrl),
	}

	go func() {
		defer close(outStream)
		defer blockingCancel()

		for {
			switch op {
			case valveOpenOP:
				select {
				case v, more := <-inStream:
					if !more {
						return
					}

					select {
					case outStream <- v:
					case <-ctx.Done():
						return
					}

				case op = <-innerValve.ctrl:

				case <-ctx.Done():
					return
				}
			case valveCloseOP:
				select {
				case <-ctx.Done():
					return
				case op = <-innerValve.ctrl:
				}

			default:
				panic("should never happen")
			}
		}
	}()

	open := innerValve.open
	_close := innerValve.close

	return outStream, open, _close
}

func (p *valve) open() {
	select {
	case p.ctrl <- valveOpenOP:
	case <-p.blockingCtx.Done():
	}
}

func (p *valve) close() {
	select {
	case p.ctrl <- valveCloseOP:
	case <-p.blockingCtx.Done():
	}
}
func RateLimit[T any](
	ctx context.Context,
	inStream <-chan T,
	ratePerSecond int,
	option ...ratelimit.Option,
) <-chan T {
	outStream := make(chan T)

	rl := ratelimit.New(ratePerSecond, option...)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case <-rl.Take(ctx):
				select {
				case val, more := <-inStream:
					if !more {
						return
					}

					select {
					case outStream <- val:
					case <-ctx.Done():
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}