package dirty

import (
	"container/heap"
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

type (
	reorderOptions struct {
		bufferSize   int
		reverseOrder bool
	}
	// ReorderOption defines options for Reorder.
	ReorderOption interface {
		apply(*reorderOptions)
	}

	reverseOrder struct{}

	bufferSize int

	revOrder[Priority dirty.Ordered, Payload any]           []dirty.Indexed[Priority, Payload]
	inOrder[Priority dirty.Ordered, Payload any]            []dirty.Indexed[Priority, Payload]
	internalHeapInterface[Index dirty.Ordered, Payload any] interface {
		heap.Interface

		NextToPop() dirty.Indexed[Index, Payload]
	}
)

const (
	defaultPrioritizeBufferSize = 10_000 // is the default buffer size
)

var (
	defaultPrioritizeOption = reorderOptions{
		// default buffer size
		bufferSize: defaultPrioritizeBufferSize,

		// not in reverse order
		reverseOrder: false,
	}

	// Check queues compatibility with heap.Interface.
	_ internalHeapInterface[int, any] = &inOrder[int, any]{}
	_ internalHeapInterface[int, any] = &revOrder[int, any]{}
)

//goland:noinspection ALL
func (q *inOrder[Priority, Payload]) NextToPop() dirty.Indexed[Priority, Payload] {
	return (*q)[0]
}

//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) NextToPop() dirty.Indexed[Priority, Payload] {
	return (*q)[0]
}

// Len implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q inOrder[Priority, Payload]) Len() int {
	return len(q)
}

// Less implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q inOrder[Priority, Payload]) Less(i, j int) bool {
	return q[i].Index < q[j].Index
}

// Swap implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q inOrder[Priority, Payload]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

// Push implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Push(x any) {
	item, _ := x.(dirty.Indexed[Priority, Payload])
	*q = append(*q, item)
}

// Pop implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *inOrder[Priority, Payload]) Pop() any {
	var zero Payload

	n := len(*q)
	item := (*q)[n-1]
	(*q)[n-1].Payload = zero // avoid memory leak
	*q = (*q)[0 : n-1]

	return item
}

// Len implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q revOrder[Priority, Payload]) Len() int {
	return len(q)
}

// Less implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q revOrder[Priority, Payload]) Less(i, j int) bool {
	return q[i].Index > q[j].Index
}

// Swap implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q revOrder[Priority, Payload]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

// Push implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Push(x any) {
	item, _ := x.(dirty.Indexed[Priority, Payload])
	*q = append(*q, item)
}

// Pop implements heap.Interface.
//
//goland:noinspection GoMixedReceiverTypes
func (q *revOrder[Priority, Payload]) Pop() any {
	var zero Payload

	n := len(*q)
	item := (*q)[n-1]
	(*q)[n-1].Payload = zero // avoid memory leak
	*q = (*q)[0 : n-1]

	return item
}

func (b bufferSize) apply(option *reorderOptions) {
	option.bufferSize = int(b)
}

// WithBufferSize sets internal buffer size.
func WithBufferSize(d int) ReorderOption {
	return bufferSize(d)
}

func (b reverseOrder) apply(option *reorderOptions) {
	option.reverseOrder = true
}

// WithOrderReversed option for reversing order.
func WithOrderReversed() ReorderOption {
	return reverseOrder{}
}

// Reorder bufferize indexed input stream.
//
//nolint:maintidx,gocognit // yes it's complex
func Reorder[Value dirty.Ordered, Payload any](
	ctx context.Context,
	inStream <-chan dirty.Indexed[Value, Payload],
	option ...ReorderOption,
) <-chan dirty.Indexed[Value, Payload] {
	outStream := make(chan dirty.Indexed[Value, Payload])

	opt := defaultPrioritizeOption
	for _, o := range option {
		o.apply(&opt)
	}

	var pq internalHeapInterface[Value, Payload]

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
				v, _ := heap.Pop(pq).(dirty.Indexed[Value, Payload])

				select {
				case outStream <- v:
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			switch {
			case pq.Len() == opt.bufferSize:
				v, _ := heap.Pop(pq).(dirty.Indexed[Value, Payload])

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
				v := pq.NextToPop()

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

func ReorderC[Value dirty.Ordered, Payload any](
	chain dirty.ChainFunc[dirty.Indexed[Value, Payload]],
	options ...ReorderOption,
) dirty.ChainFunc[dirty.Indexed[Value, Payload]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStream <-chan dirty.Indexed[Value, Payload],
	) {
		chain(ctx, wg, Reorder(ctx, inStream, options...))
	}
}
