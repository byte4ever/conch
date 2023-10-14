package conch

import (
	"context"
	"sync"
)

type (
	Key2 struct {
		A, B uint64
	}
	Hashable2 interface {
		Hash() Key2
	}
)

const (
	maxCapacity = 100
)

func Dedup2[P Hashable2, R any](
	ctx context.Context,
	inStream <-chan Request[P, R],
) <-chan Request[P, R] {
	outStream := make(chan Request[P, R])
	go func() {
		defer close(outStream)

		var m sync.Map

		valErrorChanPool := newValErrorChanPool[R](maxCapacity)
		trackerPool := newTrackerPool[R](maxCapacity)

	again:
		select {
		case <-ctx.Done():
			return
		case req, more := <-inStream:
			if !more {
				return
			}

			if dedupReturn := replicateReturnStream[R](
				trackerPool,
				valErrorChanPool,
				req.P.Hash(),
				&m,
				req.Chan,
			); dedupReturn != nil {
				select {
				case <-ctx.Done():
					return
				case outStream <- Request[P, R]{
					P:    req.P,
					Chan: dedupReturn,
				}:
					// pushed
				}
			}

		}

		goto again
	}()

	return outStream
}

func Dedup2C[P Hashable2, R any](
	chain ChainFunc[Request[P, R]],
) ChainFunc[Request[P, R]] {
	return func(
		ctx context.Context, wg *sync.WaitGroup,
		inStream <-chan Request[P, R],
	) {
		chain(ctx, wg, Dedup2(ctx, inStream))
	}
}

func replicateReturnStream[R any](
	tPool trackerPool[R],
	valErrorChanPool valErrorChanPool[R],
	key Key2,
	sMap *sync.Map,
	returnStream chan<- ValErrorPair[R],
) chan<- ValErrorPair[R] {
	pp := tPool.get()

	if nc, found := sMap.LoadOrStore(key, pp); found {
		pp.wgCollect.Done()

		c2, _ := nc.(*tracker[R])
		go c2.replicateValueIn(returnStream)

		return nil
	}

	newReturnStream := valErrorChanPool.get()

	go pp.copyFromNewReturnStream(
		returnStream,
		newReturnStream,
		key,
		sMap,
		valErrorChanPool,
	)

	return newReturnStream
}
