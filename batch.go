package conch

type (

	// IndexedRequest is a struct linking a position to a generic request.
	// V is a type constrained by the Ordered interface.
	// P represents the payload type in the request.
	// R represents the response type in the request.
	IndexedRequest[V Ordered, P, R any] struct {
		Position V
		Req      Request[P, R]
	}

	// IndexedResult represents a generic result with an associated position.
	// V is a type constrained by the Ordered interface.
	// R represents the result value type.
	IndexedResult[V Ordered, R any] struct {
		Position int
		Result   R
	}

	// IndexedResults is a slice of IndexedResult instances.
	// V represents the ordered type for indexing positions.
	// R represents the type of the result within each IndexedResult.
	IndexedResults[V Ordered, R any] []IndexedResult[V, R]

	// IndexedError associates an error with a generic positional index.
	IndexedError[V Ordered] struct {
		Position V
		Err      error
	}

	// IndexedErrors represents a slice of IndexedError elements.
	// V is a generic type constrained by the Ordered interface.
	IndexedErrors[V Ordered] []IndexedError[V]
)

func (r IndexedResults[V, R]) Len() int {
	return len(r)
}

func (r IndexedResults[V, R]) Less(i, j int) bool {
	return r[i].Position < r[j].Position
}

func (r IndexedResults[V, R]) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (e IndexedErrors[V]) Len() int {
	return len(e)
}

func (e IndexedErrors[V]) Less(i, j int) bool {
	return e[i].Position < e[j].Position
}

func (e IndexedErrors[V]) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

/*
func Batcher[P, R any](
	ctx context.Context,
	input []P,
	requestFunc RequestProcessingFunc[P, R],
	concurrency int,
) (IndexedResults[int, R], IndexedErrors[int, R]) {
	var wg *sync.WaitGroup

	f := RequesterC(ctx, wg,
		MultiplexC[Request[P, R]](
			concurrency,
			RequestConsumersC[P, R](
				requestFunc,
			),
		),
	)

	for idx, p := range input {
		x := f(ctx, p)
	}

	f(ctx, input...)
}
*/
