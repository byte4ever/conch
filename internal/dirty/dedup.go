package dirty

type Key struct {
	A, B uint64
}

type Hashable interface {
	Hash() Key
}

/*import (
	"context"
	"sync"
)

type Key struct {
	A, B uint64
}

type Hashable interface {
	Hash() Key
}

func Dedup[P Hashable, R any](
	ctx context.Context,
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])
	go func() {
		defer close(outStream)

		dedupPool := NewDedupPool[R](100)

		defer dedupPool.Purge()

		var m sync.Map

	again:
		select {
		case <-ctx.Done():
			return
		case req, more := <-inStream:
			if !more {
				return
			}

			if dedupReturn := pushFunc(
				dedupPool,
				req.P.Hash(),
				&m,
				req.Return,
			); dedupReturn != nil {
				select {
				case <-ctx.Done():
					return
				case outStream <- Request[P, R]{
					P:      req.P,
					Return: dedupReturn,
				}:
					// pushed
				}
			}

		}

		goto again
	}()

	return outStream
}

func DedupC[P Hashable, R any](
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, Dedup(ctx, inStream))
	}
}

type PP[T any] struct {
	mtx *sync.Mutex
	wg  sync.WaitGroup
	val T
	ctx context.Context
}

func pushFunc[R any](
	dedupPool DedupPool[R],
	k Key,
	m *sync.Map,
	f ReturnFun[R],
) ReturnFun[R] {
	// todo :- lmartin 5/15/23 -: use pool to avoid memory allocation

	c := dedupPool.Get()
	c.mtx.Lock()

	if nc, found := m.LoadOrStore(k, c); found {
		// be very careful here need to unlock unused structure before
		// putting it back to pool
		c.mtx.Unlock()
		dedupPool.PushBack(c)

		c2, _ := nc.(*PP[ValErrorPair[R]])

		c2.wg.Add(1)

		go func() {
			defer c2.wg.Done()
			c2.mtx.Lock()
			c2.mtx.Unlock()
			f(c2.ctx, c2.val)
		}()

		return nil
	}

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		c.mtx.Lock()
		c.mtx.Unlock()
		f(c.ctx, c.val)
	}()

	return func(ctx context.Context, v ValErrorPair[R]) {
		m.Delete(k)

		c.ctx = ctx
		c.val = v
		c.mtx.Unlock()
		c.wg.Wait()
		dedupPool.PushBack(c)
	}
}

type DedupPool[R any] chan *PP[ValErrorPair[R]]

func NewDedupPool[R any](maxSize int) DedupPool[R] {
	return make(DedupPool[R], maxSize)
}

func (p DedupPool[R]) Get() *PP[ValErrorPair[R]] {
	select {
	case v := <-p:
		return v
	default:
		return &PP[ValErrorPair[R]]{
			mtx: &sync.Mutex{},
			wg:  sync.WaitGroup{},
			ctx: nil,
		}
	}
}

func (p DedupPool[R]) PushBack(v *PP[ValErrorPair[R]]) {
	select {
	case p <- v:
	default:
		// trashIt
		v.mtx = nil
		v.ctx = nil
	}
}

func (p DedupPool[R]) Purge() {
	close(p)

	for p2 := range p {
		// trashIt
		p2.mtx = nil
		p2.ctx = nil
		p2 = nil
	}
}
*/
