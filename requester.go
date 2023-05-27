package conch

import (
	"context"
	"sync"
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
	wg *sync.WaitGroup,
	count int,
) (
	[]RequestFunc[P, R],
	<-chan Request[P, R],
) {
	prioRequesters := make(
		[]RequestFunc[P, R],
		count,
	)

	streams := make(
		[]<-chan Request[P, R],
		count,
	)

	for i := 0; i < count; i++ {
		prioRequesters[i], streams[i] = Requester[P, R](ctx, wg)
	}

	return prioRequesters, UnfairFanIn(ctx, streams...)
}

// UnfairRequestersC is the chained version of UnfairRequesters.
func UnfairRequestersC[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	chain ChainFunc[Request[P, R]],
) []RequestFunc[P, R] {
	requesters, o := UnfairRequesters[P, R](ctx, wg, count)

	chain(ctx, wg, o)

	return requesters
}

func Requester[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
) (
	RequestFunc[P, R],
	<-chan Request[P, R],
) {
	outStream := make(chan Request[P, R])

	go func() {
		defer func() {
			wg.Wait()
			close(outStream)
			for r := range outStream {
				r.Return(
					ctx,
					ValErrorPair[R]{
						Err: ctx.Err(),
					},
				)
			}
		}()

		select {
		case <-ctx.Done():
			return
		}
	}()

	return requestFun[P, R](ctx, wg, outStream),
		outStream
}

func requestFun[P, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	outStream chan Request[P, R],
) RequestFunc[P, R] {
	return func(
		innerCtx context.Context,
		params P,
	) (
		R,
		error,
	) {
		var zeroR R

		chResp := make(chan ValErrorPair[R], 1)

		if innerCtx.Err() != nil {
			return zeroR, ctx.Err()
		}

		if ctx.Err() != nil {
			return zeroR, ctx.Err()
		}

		req := Request[P, R]{
			P: params,
			Return: func(ctx context.Context, v ValErrorPair[R]) {
				defer close(chResp)

				select {
				case <-ctx.Done():
					return
				case chResp <- v:
				}
			},
		}

		// send request
		wg.Add(1)
		select {
		case outStream <- req:
		case <-ctx.Done():
			wg.Done()
			return zeroR, ctx.Err()
		case <-innerCtx.Done():
			wg.Done()
			return zeroR, innerCtx.Err()
		}

		wg.Done()

		// receive response
		select {
		case rv := <-chResp:
			return rv.V, rv.Err
		case <-ctx.Done():
			return zeroR, ctx.Err()
		case <-innerCtx.Done():
			return zeroR, innerCtx.Err()
		}
	}
}

func RequesterC[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	chain ChainFunc[Request[P, R]],
) RequestFunc[P, R] {
	requester, s := Requester[P, R](ctx, wg)
	chain(ctx, wg, s)

	return requester
}
