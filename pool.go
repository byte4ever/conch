package conch

type (
	// Pool is a generic structure representing a resource pool. Pool manages
	// resources of any type R that implements the Cleanable interface. R
	// instances are retrieved from or returned to the channel C. Pool uses the
	// Producer function to generate new instances of R when needed.
	Pool[R any] chan *R

	// PoolGetter is an interface for retrieving instances of type R. R must
	// satisfy the Cleanable interface. The Get method fetches an instance of R
	// from the pool or creates one.
	PoolGetter[R any] interface {
		Get() *R
	}

	// PoolReceiver defines an interface for returning objects to a pool. R
	// represents any type that can be handled by the implementing type. The
	// PutBack method is used to return an object to the pool.
	PoolReceiver[T any] interface {
		PutBack(p *T)
	}

	// PoolProvider represents an interface that combines PoolGetter and
	// PoolReceiver.
	PoolProvider[R any] interface {
		PoolGetter[R]
		PoolReceiver[R]
	}
)

// NewPool initializes a Pool with a maximum capacity and a producer function.
// maxCapacity determines the pool's size and producer creates new Cleanable
// resources. Returns a pointer to the newly created Pool.
func NewPool[R any](
	maxCapacity int,
) Pool[R] {
	return make(
		chan *R,
		maxCapacity,
	)
}

// Get retrieves a resource of type R from the pool, creating one if needed.
func (v Pool[R]) Get() (c *R) {
	select {
	case c = <-v:
		return
	default:
		return new(R)
	}
}

// PutBack returns a cleaned resource to the pool for reuse. If the pool's
// channel is full, the resource is discarded.
func (v Pool[R]) PutBack(p *R) {
	var zero R
	*p = zero

	select {
	case v <- p:
		// store to reuse it
	default:
		// trash it
	}
}
