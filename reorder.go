package conch

import (
	"container/heap"
	"context"
	"time"
)

type (
	prioritizeOption struct {
		backPressureDelay time.Duration
		flushWhenDown     bool
		bufferSize        int
		reverseOrder      bool
	}

	// ReorderOption defines options for Reorder.
	ReorderOption interface {
		apply(*prioritizeOption)
	}

	backPressureDelayOption time.Duration

	reverseOrder struct{}

	bufferSize int

	revOrder[Priority Ordered, Payload any] []Indexed[Priority, Payload]

	inOrder[Priority Ordered, Payload any] []Indexed[Priority, Payload]
)

func (q *inOrder[Priority, Payload]) ToPop() Indexed[Priority, Payload] {
	return (*q)[0]
}

func (q *revOrder[Priority, Payload]) ToPop() Indexed[Priority, Payload] {
	return (*q)[0]
}

// Len implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Len() int {
	return len(*q)
}

// Less implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Less(i, j int) bool {
	return (*q)[i].Index < (*q)[j].Index
}

// Swap implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Swap(i, j int) {
	(*q)[i], (*q)[j] = (*q)[j], (*q)[i]
}

// Push implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Push(x any) {
	item, _ := x.(Indexed[Priority, Payload])
	*q = append(*q, item)
}

// Pop implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = Indexed[Priority, Payload]{} // avoid memory leak
	*q = old[0 : n-1]

	return item
}

// Len implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Len() int {
	return len(*q)
}

// Less implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Less(i, j int) bool {
	return (*q)[i].Index > (*q)[j].Index
}

// Swap implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Swap(i, j int) {
	(*q)[i], (*q)[j] = (*q)[j], (*q)[i]
}

// Push implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Push(x any) {
	item, _ := x.(Indexed[Priority, Payload])
	*q = append(*q, item)
}

// Pop implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = Indexed[Priority, Payload]{} // avoid memory leak
	*q = old[0 : n-1]

	return item
}

const (
	defaultPrioritizeBufferSize = 1000 // is the default buffer size
)

var (
	defaultPrioritizeOption = prioritizeOption{
		// no back pressure delay
		backPressureDelay: 0,

		// buffer size is 100
		bufferSize: defaultPrioritizeBufferSize,

		// not in reverse order
		reverseOrder: false,
	}

	// Check queues compatibility with heap.Interface.
	_ MyI[int, any] = &inOrder[int, any]{}
	_ MyI[int, any] = &revOrder[int, any]{}
)

func (b backPressureDelayOption) apply(option *prioritizeOption) {
	option.backPressureDelay = time.Duration(b)
}

// WithBackPressureDelay force internal back pressure on stream by inducing
// some delay latency  in reordering.
// This helps to fill internal buffer when using reordering as a priority
// stream.
func WithBackPressureDelay(d time.Duration) ReorderOption {
	return backPressureDelayOption(d)
}

func (b bufferSize) apply(option *prioritizeOption) {
	option.bufferSize = int(b)
}

// WithBufferSize sets internal buffer size.
func WithBufferSize(d int) ReorderOption {
	return bufferSize(d)
}

func (b reverseOrder) apply(option *prioritizeOption) {
	option.reverseOrder = true
}

// WithOrderReversed option for reversing order.
func WithOrderReversed() ReorderOption {
	return reverseOrder{}
}

type MyI[Index Ordered, Payload any] interface {
	heap.Interface
	ToPop() Indexed[Index, Payload]
}

// Reorder bufferize indexed input stream.
//
//nolint:maintidx,gocognit // yes it's complex
func Reorder[Value Ordered, Payload any](
	ctx context.Context,
	inStream <-chan Indexed[Value, Payload],
	option ...ReorderOption,
) <-chan Indexed[Value, Payload] {

	outStream := make(chan Indexed[Value, Payload])

	opt := defaultPrioritizeOption
	for _, o := range option {
		o.apply(&opt)
	}

	var pq MyI[Value, Payload]

	if opt.reverseOrder {
		l := make(revOrder[Value, Payload], 0, opt.bufferSize)
		pq = &l
	} else {
		l := make(inOrder[Value, Payload], 0, opt.bufferSize)
		pq = &l
	}

	// reader
	go func() {
		defer close(outStream)
		defer func() {
			for ctx.Err() == nil && pq.Len() > 0 {
				v := heap.Pop(pq).(Indexed[Value, Payload])

				select {
				case outStream <- v:
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			// fmt.Println(pq)
			switch {
			case pq.Len() == opt.bufferSize:
				v := heap.Pop(pq).(Indexed[Value, Payload])
				select {
				case outStream <- v:
				case <-ctx.Done():
					return
				}

			case pq.Len() == 0:
				select {
				case v, more := <-inStream:
					if !more {
						return
					}
					heap.Push(pq, v)

				case <-ctx.Done():
					return
				}
			default:
				v := pq.ToPop()

				select {
				case v2, more := <-inStream:
					if !more {
						return
					}

					heap.Push(pq, v2)

				case outStream <- v:
					heap.Pop(pq)

				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}
