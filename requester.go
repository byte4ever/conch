package conch

import (
	"context"
	"fmt"
	"sync"
)

type Request[P any, R any] struct {
	P      P
	ChResp chan<- ValErrorPair[R]
}

func Requester[P, R any](ctx context.Context) (
	func(
		context.Context,
		P,
	) (
		R,
		error,
	),
	<-chan Request[P, R],
) {
	outStream := make(chan Request[P, R])

	go func() {
		defer close(outStream)
		select {
		case <-ctx.Done():
			return
		}
	}()

	return func(
		innerCtx context.Context,
		params P,
	) (
		R,
		error,
	) {
		var zeroR R

		chResp := make(chan ValErrorPair[R])

		if innerCtx.Err() != nil {
			return zeroR, ctx.Err()
		}

		req := Request[P, R]{
			P:      params,
			ChResp: chResp,
		}

		select {
		case outStream <- req:
		case <-ctx.Done():
			return zeroR, ctx.Err()
		case <-innerCtx.Done():
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
	}, outStream
}

func SpawnRequestProcessor[P any, R any](
	ctx context.Context, inStream <-chan Request[P, R],
	processing func(context.Context, P) (R, error), id string,
) func() {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for req := range inStream {
			// fmt.Println("processing    --- ", id, req.P)
			result, err := processing(ctx, req.P)
			select {
			case req.ChResp <- ValErrorPair[R]{V: result, Err: err}:
				close(req.ChResp)
			case <-ctx.Done():
				return
			}
		}
	}()

	return wg.Wait
}

func SpawnRequestProcessorsPool[P any, R any](
	ctx context.Context, inStream <-chan Request[P, R],
	processing func(context.Context, P) (R, error), count int, id string,
) func() {
	f := make([]func(), count)

	for i := 0; i < count; i++ {
		f[i] = SpawnRequestProcessor(
			ctx,
			inStream,
			processing,
			fmt.Sprintf("%s-%d", id, i),
		)
	}

	return func() {
		for _, f := range f {
			f()
		}
	}
}
