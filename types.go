package conch

import (
	"context"
	"sync"
)

type (
	// Signed is a constraint that permits any signed integer type.
	// If future releases of Go add new predeclared signed integer types,
	// this constraint will be modified to include them.
	Signed interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}

	// Unsigned is a constraint that permits any unsigned integer type.
	// If future releases of Go add new predeclared unsigned integer types,
	// this constraint will be modified to include them.
	Unsigned interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
	}

	// Integer is a constraint that permits any integer type.
	// If future releases of Go add new predeclared integer types,
	// this constraint will be modified to include them.
	Integer interface {
		Signed | Unsigned
	}

	// Float is a constraint that permits any floating-point type.
	// If future releases of Go add new predeclared floating-point types,
	// this constraint will be modified to include them.
	Float interface {
		~float32 | ~float64
	}

	// Ordered is a constraint that permits any ordered type: any type
	// that supports the operators < <= >= >.
	// If future releases of Go add new ordered types,
	// this constraint will be modified to include them.
	Ordered interface {
		Integer | Float | ~string
	}

	// Indexed represents a payload with an ordered index.
	// It combines an index value with a payload, commonly used for
	// maintaining order in stream processing operations.
	Indexed[V Ordered, Payload any] struct {
		Index   V       // The index value used for ordering
		Payload Payload // The actual data payload
	}

	// IndexedInteger represents a payload with an integer index.
	// Similar to Indexed but specifically constrained to integer index types.
	IndexedInteger[V Integer, Payload any] struct {
		Index   V       // The integer index value
		Payload Payload // The actual data payload
	}

	Generator[T any] func(ctx context.Context) (output <-chan T, err error)

	Doer[T any] func(ctx context.Context, id int, param T)

	// Processor defines a function that read from a single input stream and
	// produce elements to the resulting output stream.
	// Processor[From, To any] func(
	// 	ctx context.Context, input <-chan From,
	// ) <-chan To

	// ValErrorPair represents a value-error pair for result handling.
	// It encapsulates either a successful value or an error state,
	// commonly used in stream processing for error propagation.
	ValErrorPair[V any] struct {
		V   V     // The value component (zero value if error occurred)
		Err error // The error component (nil if successful)
	}

	FilterFunc[T any] func(ctx context.Context, v T) bool

	ChainFunc[T any] func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan T,
	)

	ChainsFunc[T any] func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream ...<-chan T,
	)

	// Request represents a request with parameters and response channel.
	// It encapsulates the parameter data, context, and a channel for receiving
	// the response wrapped in a ValErrorPair for error handling.
	Request[P any, R any] struct {
		P    P                        // The request parameters
		Ctx  context.Context          // The request context for cancellation
		Chan chan<- ValErrorPair[R]   // Channel for receiving the response
	}

	ReturnFun[R any] func(context.Context, ValErrorPair[R])

	IndexedRequestFunc[V Ordered, P, R any] func(
		ctx context.Context,
		params Indexed[V, P],
	) (
		Indexed[V, R],
		error,
	)

	RequestFunc[P, R any] func(
		ctx context.Context,
		params P,
	) (
		R,
		error,
	)

	RequestProcessingFunc[P, R any] func(
		ctx context.Context,
		id int,
		params P,
	) (
		R,
		error,
	)

	ValErrorPairProvider[R any] func(ctx context.Context) ValErrorPair[R]

	// None represents an empty struct used as a placeholder or unit type.
	// It's commonly used when a type parameter is required but no actual data is needed.
	None struct{}
)

func ToValError[V any](v V, err error) ValErrorPair[V] {
	return ValErrorPair[V]{
		V:   v,
		Err: err,
	}
}

func WrapToIndexedRequester[V Ordered, P, R any](
	f RequestFunc[P, R],
) IndexedRequestFunc[V, P, R] {
	return func(
		ctx context.Context,
		params Indexed[V, P],
	) (Indexed[V, R], error) {
		result, err := f(ctx, params.Payload)

		return Indexed[V, R]{
			Index:   params.Index,
			Payload: result,
		}, err
	}
}

func (p ValErrorPair[V]) SetValue(v V) {
	p.V, p.Err = v, nil
}

func (p ValErrorPair[V]) SetError(e error) {
	var zero V
	p.V, p.Err = zero, e
}
