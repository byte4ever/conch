package dirty

import (
	"context"
)

// ContextBreaker creates an output stream that copy input stream and
// that closeMe when the context is done or input stream is closed.
//
//	This is useful to cut properly stream flows, especially when down stream
//	enter some operators that are no longer sensitive to context termination.
func ContextBreaker[T any](
	ctx context.Context,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T)

	go func() {
		defer close(outStream)

		for t := range inStream {
			select {
			case <-ctx.Done():
				return

			case outStream <- t:
			}
		}
	}()

	return outStream
}

// ContextBreakerBuffer creates an output buffered stream that copy input stream
// and that close when the context is done or input stream is closed.
//
// Buffer size is defined by bufferSize parameter.
//
//	This is useful to cut properly stream flows, especially when down stream
//	enter some operators that are no longer sensitive to context termination.
func ContextBreakerBuffer[T any](
	ctx context.Context,
	bufferSize int,
	inStream <-chan T,
) <-chan T {
	outStream := make(chan T, bufferSize)

	go func() {
		defer close(outStream)

		for t := range inStream {
			select {
			case <-ctx.Done():
				return

			case outStream <- t:
			}
		}
	}()

	return outStream
}
