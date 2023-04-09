package conch

import (
	"context"
)

func Transform[In any, Out any](
	ctx context.Context,
	inStream <-chan In,
	transformer func(ctx context.Context, in In) Out,
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
