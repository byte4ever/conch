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
	// Complex is a constraint that permits any complex numeric type.
	// If future releases of Go add new predeclared complex numeric types,
	// this constraint will be modified to include them.
	Complex interface {
		~complex64 | ~complex128
	}
	// Ordered is a constraint that permits any ordered type: any type
	// that supports the operators < <= >= >.
	// If future releases of Go add new ordered types,
	// this constraint will be modified to include them.
	Ordered interface {
		Integer | Float | ~string
	}
	Indexed[V Ordered, Payload any] struct {
		Index   V
		Payload Payload
	}
	IndexedInteger[V Integer, Payload any] struct {
		Index   V
		Payload Payload
	}
	Comparable interface {
		LessThan(other Comparable) bool
	}
	Generator[T any] func(context.Context) (output <-chan T, err error)
	Doer[T any]      func(context.Context, T)
	// Processor defines a function that read from a single input stream and
	// produce elements to the resulting output stream.
	Processor[From, To any] func(
		ctx context.Context, input <-chan From,
	) <-chan To
	ValErrorPair[V any] struct {
		V   V
		Err error
	}
	FilterFunc[T any] func(ctx context.Context, v T) bool
	ChainFunc[T any]  func(
		ctx context.Context,
		group *sync.WaitGroup,
		inStream <-chan T,
	)
	ChainsFunc[T any] func(
		ctx context.Context,
		group *sync.WaitGroup,
		inStream ...<-chan T,
	)
	Request[P any, R any] struct {
		P      P
		Return func(context.Context, ValErrorPair[R])
	}
	ReturnFun[R any]      func(context.Context, ValErrorPair[R])
	RequestFunc[P, R any] func(
		context.Context,
		P,
	) (
		R,
		error,
	)
)
