package conch

import (
	"context"
	"sync"
)

func Requester[P, R any](
	ctx context.Context,
) (
	RequestFunc[P, R],
	<-chan Request[P, R],
) {
	var wg sync.WaitGroup

	outStream := make(chan Request[P, R])

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

	chanPool := newValErrorChanPool[R](100)

	return func(
			innerCtx context.Context,
			params P,
		) (
			R,
			error,
		) {
			wg.Add(1)
			defer wg.Done()

			chResp := chanPool.get()
			defer chanPool.putBack(chResp)

			var zeroR R

			select {
			case outStream <- Request[P, R]{
				P:    params,
				Ctx:  innerCtx,
				Chan: chResp,
			}:
				select {
				case rv := <-chResp:
					return rv.V, rv.Err

				case <-ctx.Done():
					return zeroR, ctx.Err()

				case <-innerCtx.Done():
					return zeroR, innerCtx.Err()
				}

			case <-ctx.Done():
				return zeroR, ctx.Err()

			case <-innerCtx.Done():
				return zeroR, innerCtx.Err()
			}
		},

		outStream
}

func RequesterC[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	chains ChainsFunc[Request[P, R]],
) RequestFunc[P, R] {
	requester, s := Requester[P, R](ctx)
	chains(ctx, wg, s)

	return requester
}

func RequestersC[P any, R any](
	ctx context.Context,
	wg *sync.WaitGroup,
	count int,
	chains ChainsFunc[Request[P, R]],
) []RequestFunc[P, R] {
	requesters := make([]RequestFunc[P, R], count)
	streams := make([]<-chan Request[P, R], count)

	for i := 0; i < count; i++ {
		requesters[i], streams[i] = Requester[P, R](ctx)
	}

	chains(ctx, wg, streams...)

	return requesters
}
