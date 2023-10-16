package dirty

import (
	"context"
	"sync"

	"github.com/byte4ever/conch"
)

func Transform[In any, Out any](
	ctx context.Context,
	transformer func(ctx context.Context, in In) Out,
	inStream <-chan In,
) <-chan Out {
	outStream := make(chan Out)

	go func() {
		defer close(outStream)

		for {
			select {
			case <-ctx.Done():
				return

			case in, more := <-inStream:
				if !more {
					return
				}

				o := transformer(ctx, in)

				select {
				case outStream <- o:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return outStream
}

func TransformC[In any, Out any](
	transformer func(ctx context.Context, in In) Out,
	chain dirty.ChainFunc[Out],
) dirty.ChainFunc[In] {
	return func(
		ctx context.Context,
		wg *sync.WaitGroup,
		inStream <-chan In,
	) {
		s := Transform(ctx, transformer, inStream)
		chain(ctx, wg, s)
	}
}
