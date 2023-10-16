package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

// UnfairRequesters returns a list of thread safe requester functions in
// priority order. Priority order follows index in the slice.
//
// This mean first requester in the slice got the lowest priority and the latest
// requester in the slice got the highest one.
//
// Keep in mind that priority effect is only achieved under high pressure.
func UnfairRequesters[P, R any](
	ctx context.Context,
	count int,
) (
	[]dirty.RequestFunc[P, R],
	<-chan dirty.Request[P, R],
) {
	prioRequesters := make(
		[]dirty.RequestFunc[P, R],
		count,
	)

	streams := make(
		[]<-chan dirty.Request[P, R],
		count,
	)

	for i := 0; i < count; i++ {
		prioRequesters[i], streams[i] = Requester[P, R](ctx)
	}

	return prioRequesters, UnfairFanIn(ctx, streams...)
}

// UnfairRequestersC is the chained version of UnfairRequesters.
func UnfairRequestersC[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	chain dirty.ChainFunc[dirty.Request[P, R]],
) []dirty.RequestFunc[P, R] {
	requesters, o := UnfairRequesters[P, R](ctx, count)

	chain(ctx, wg, o)

	return requesters
}

func Requester[P, R any](
	ctx context.Context,
) (
	dirty.RequestFunc[P, R],
	<-chan dirty.Request[P, R],
) {
	var wg sync.WaitGroup

	outStream := make(chan dirty.Request[P, R])

	go func() {
		defer func() {
			wg.Wait()
			close(outStream)
		}()

		select {
		case <-ctx.Done():
			return
		}
	}()

	chanPool := newValErrorChanPool[R](10)

	return func(
			innerCtx context.Context,
			params P,
		) (
			R,
			error,
		) {
			var zeroR R

			chResp := chanPool.get()
			defer chanPool.putBack(chResp)

			select {
			case <-innerCtx.Done():
				return zeroR, innerCtx.Err()
			case <-ctx.Done():
				return zeroR, ctx.Err()
			default:
				// send request
				wg.Add(1)
				select {
				case outStream <- dirty.Request[P, R]{
					P:    params,
					Ctx:  innerCtx,
					Chan: chResp,
				}:
					wg.Done()
				case <-ctx.Done():
					wg.Done()
					return zeroR, ctx.Err()
				case <-innerCtx.Done():
					wg.Done()
					return zeroR, innerCtx.Err()
				}

				select {
				case rv := <-chResp:
					return rv.V, rv.Err
				case <-ctx.Done():
					return zeroR, ctx.Err()
				case <-innerCtx.Done():
					return zeroR, innerCtx.Err()
				}
			}
		},
		outStream
}

func RequesterC[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	chain dirty.ChainFunc[dirty.Request[P, R]],
) dirty.RequestFunc[P, R] {
	requester, s := Requester[P, R](ctx)
	chain(ctx, wg, s)

	return requester
}
